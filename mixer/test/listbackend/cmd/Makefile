.PHONY: deploy docker

listadapter: main.go
	../../../../bin/gobuild.sh listadapter istio.io/istio/pkg/version .
docker: listadapter Dockerfile
	docker build --build-arg https_proxy=http://15.85.195.199:8080/ --build-arg http_proxy=http://15.85.195.199:8080/ . -t gcr.io/mixologist-142215/listadapter:latest

deploy:
	kubectl -n istio-system apply -f listadapter.yaml
