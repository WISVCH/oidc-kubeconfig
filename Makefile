.PHONY: all build push

TAG:=dev-$(shell date +%s)

all: build push

build:
	@docker build --no-cache --pull -t wisvch/oidc-kubeconfig:${TAG} .

push:
	@docker push wisvch/oidc-kubeconfig:${TAG}
