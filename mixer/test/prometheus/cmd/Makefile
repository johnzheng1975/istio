.PHONY: deploy docker

prometheusadapter: main.go
	../../../../bin/gobuild.sh prometheusadapter istio.io/istio/pkg/version .
docker: prometheusadapter
	docker build --build-arg https_proxy=http://15.85.195.199:8080/ --build-arg http_proxy=http://15.85.195.199:8080/ . -t gcr.io/mixologist-142215/prometheusadapter:latest

deploy:
	kubectl -n istio-system apply -f prometheusadapter.yaml
