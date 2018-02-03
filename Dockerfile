FROM golang AS builder
COPY . /go/src/github.com/wisvch/oidc-kubeconfig
RUN go install github.com/wisvch/oidc-kubeconfig

FROM wisvch/debian:stretch-slim
WORKDIR /srv
COPY --from=builder /go/bin/oidc-kubeconfig /srv
COPY template.html /srv
ENTRYPOINT ["/srv/oidc-kubeconfig"]
