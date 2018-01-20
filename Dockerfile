FROM golang AS builder
COPY . /go/src/github.com/wisvch/oidc-kubeconfig
RUN go install github.com/wisvch/oidc-kubeconfig

FROM wisvch/debian:stretch
COPY --from=builder /go/bin/oidc-kubeconfig /usr/local/bin
ENTRYPOINT ["/usr/local/bin/oidc-kubeconfig"]
