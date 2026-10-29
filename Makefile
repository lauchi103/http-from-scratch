APP_NAME      ?= mein-httpserver
TAG           ?= latest
IMAGE         ?= $(APP_NAME):$(TAG)
NAMESPACE     ?= default
MANIFEST      ?=http-server.yaml
KIND_CLUSTER  ?= kind

all: build deploy

build:
	docker build -t $(IMAGE) .

deploy:
	kubectl apply -f $(MANIFEST) -n $(NAMESPACE)
