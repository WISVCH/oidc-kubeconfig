FROM golang AS builder
WORKDIR /src
COPY . .
RUN go install

FROM wisvch/debian:stretch-slim
WORKDIR /srv
COPY --from=builder /go/bin/oidc-kubeconfig /srv
COPY template.html /srv
RUN groupadd -r oidc-kubeconfig --gid=999 && useradd --no-log-init -r -g oidc-kubeconfig --uid=999 oidc-kubeconfig
USER 999
ENTRYPOINT ["/srv/oidc-kubeconfig"]
