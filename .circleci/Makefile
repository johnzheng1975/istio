# Builder version
CI_VERSION ?= go1.10-k8s1.9-helm2.7.2-minikube0.25
CI_HUB ?= istio

ci.image:
	docker build --build-arg https_proxy=http://15.85.195.199:8080/ --build-arg http_proxy=http://15.85.195.199:8080/ -t ${CI_HUB}/ci:$(CI_VERSION) -f Dockerfile .

ci.push:
	docker push "${CI_HUB}/ci:$(CI_VERSION)"

ci.run:
	docker run --rm -u $(shell id -u) -it \
        -v ${GOPATH}:${GOPATH} \
        -w ${PWD} \
        -e USER=${USER} \
		--entrypoint /bin/bash \
        ${CI_HUB}/ci:$(CI_VERSION)


.PHONY: ci.image ci.push ci.run
