.PHONY: all build push

TAG:=dev-$(shell date +%s)

all: build push

build:
	@docker build --no-cache --pull -t quay.io/wisvch/oidc-kubeconfig:${TAG} .

push:
	@docker push quay.io/wisvch/oidc-kubeconfig:${TAG}
