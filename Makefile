APP_NAME      ?= mein-httpserver
TAG           ?= latest
IMAGE         ?= $(APP_NAME):$(TAG)
NAMESPACE     ?= default
MANIFEST      ?=http-server.yaml
KIND_CLUSTER  ?= kind
LOCAL_PORT    ?= 8080
SVC_PORT      ?= 8080
PID_FILE      ?= .port-forward.pid

all: build deploy

build:
	docker build -t $(IMAGE) .

deploy:
	kubectl apply -f $(MANIFEST) -n $(NAMESPACE)
port-forward: stop-forward
	@echo "Starte Port-Forward auf localhost: $(LOCAL_PORT)..."
	@kubectl port-forward -n $(NAMESPACE) svc/httpserver-svc $(LOCAL_PORT):$(SVC_PORT) > /dev/null 2>&1 & echo $$! > $(PID_FILE)
	@sleep 1
	@echo "Bereit! Service erreichbar unter http://localhost:$(LOCAL_PORT)"

stop-forward:
	@if [ -f $(PID_FILE) ]; then \
		kill -9 $$(cat $(PID_FILE)) > /dev/null 2>&1 || true; \
		rm -f $(PID_FILE); \
		echo "Port-Forward beendet."; \
	fi