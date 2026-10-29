APP_NAME      ?= mein-httpserver
TAG           ?= latest
IMAGE         ?= $(APP_NAME):$(TAG)
NAMESPACE     ?= default
MANIFEST      ?=http-server.yaml
KIND_CLUSTER  ?= kind
LOCAL_PORT    ?= 8080
SVC_PORT      ?= 8080
PID_FILE      ?= .port-forward.pid
VM_PORT       ?= 8428
MON_NAMESPACE ?= monitoring
GRAFANA_PORT  ?= 3000

all: build deploy

build:
	docker build -t $(IMAGE) .

deploy:
	kubectl apply -f $(MANIFEST) -n $(NAMESPACE)
port-forward: stop-forward
	@echo "Starte Port-Forward auf localhost: $(LOCAL_PORT)..."
	@kubectl port-forward -n $(NAMESPACE) svc/httpserver-svc $(LOCAL_PORT):$(SVC_PORT) > /dev/null 2>&1 & echo $$! >> $(PID_FILE)
	@kubectl port-forward -n $(MON_NAMESPACE) svc/vmsingle-vm-stack-victoria-metrics-k8s-stack $(VM_PORT):$(VM_PORT) > /dev/null 2>&1 & echo $$! >> $(PID_FILE)
	@kubectl port-forward -n $(MON_NAMESPACE) vm-stack-grafana-6fc8cc8c5b-49fks $(GRAFANA_PORT):$(GRAFANA_PORT) > /dev/null 2>&1 & echo $$! >> $(PID_FILE)
	@sleep 1
	@echo "Bereit! Service erreichbar unter http://localhost:$(LOCAL_PORT)"
	@echo "Bereit! VMUI erreichbar unter http://localhost:$(VM_PORT)/vmui"

stop-forward:
	@if [ -f $(PID_FILE) ]; then \
		echo "Beende bestehende Port-Forwards..."; \
		while read -r pid; do \
			kill -9 $$pid > /dev/null 2>&1 || true; \
		done < $(PID_FILE); \
		rm -f $(PID_FILE); \
	fi