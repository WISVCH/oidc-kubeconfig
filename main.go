package main

import (
	"fmt"
	"github.com/coreos/go-oidc"
	"github.com/dchest/uniuri"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"html/template"
	"log"
	"net/http"
)

type Config struct {
	ClientID     string `required:"true"`
	ClientSecret string `required:"true"`
	Issuer       string `required:"true"`
	RedirectURL  string `required:"true"`
}

type KubeAuthData struct {
	ClientID     string
	ClientSecret string
	Issuer       string
	Subject      string
	RefreshToken string
}

var (
	sessionName = "kubeauth"

	store = sessions.NewCookieStore(
		securecookie.GenerateRandomKey(32),
		securecookie.GenerateRandomKey(32),
	)
)

func init() {
	store.Options.MaxAge = 0
}

func main() {
	var config Config
	err := envconfig.Process("kubeconfig", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, config.Issuer)
	if err != nil {
		log.Fatal(err.Error())
	}
	provider.Endpoint()
	oauth2Config := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       []string{"ldap", "openid", "offline_access"},
		Endpoint:     provider.Endpoint(),
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: config.ClientID})

	tmpl := template.Must(template.ParseFiles("template.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		session, _ := store.Get(r, sessionName)
		val, ok := session.Values["subject"]
		if !ok {
			state := uniuri.New()
			session.Values["state"] = state
			session.Save(r, w)
			http.Redirect(w, r, oauth2Config.AuthCodeURL(state), http.StatusFound)
			return
		}
		subject := val.(string)
		refreshToken := session.Values["refreshToken"].(string)

		kubeAuthData := KubeAuthData{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Issuer:       config.Issuer,
			Subject:      subject,
			RefreshToken: refreshToken,
		}

		tmpl.Execute(w, kubeAuthData)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, sessionName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		val, ok := session.Values["state"]
		if !ok || r.URL.Query().Get("state") != val.(string) {
			http.Error(w, "State did not match", http.StatusBadRequest)
			return
		}

		oauth2Token, err := oauth2Config.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			http.Error(w, "No ID token", http.StatusInternalServerError)
			return
		}

		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		session.Values["subject"] = idToken.Subject
		session.Values["refreshToken"] = oauth2Token.RefreshToken
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusFound)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	log.Printf("listening on http://%s/", "0.0.0.0:8000")
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}
