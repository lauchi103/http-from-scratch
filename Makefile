APP_NAME       ?= mein-httpserver
IMAGE          ?= $(APP_NAME)
NAMESPACE      ?= default
MANIFEST       ?= http-server.yaml
DEPLOYMENT_NAME?= httpserver
LOCAL_PORT     ?= 8080
SVC_PORT       ?= 8080
PID_FILE       ?= .port-forward.pid
VM_PORT        ?= 8428
MON_NAMESPACE  ?= monitoring
GRAFANA_PORT   ?= 3000

.PHONY: all build deploy reload port-forward stop-forward

all: build deploy

## build: Baut das Docker Image mit eindeutigem Versions-Tag
build:
	@echo "Baue Docker Image $(IMAGE)..."
	docker build -t $(IMAGE) .

## deploy: Wendet Manifeste an, setzt das neue Image und wartet auf den Rollout
deploy:
	kubectl apply -f $(MANIFEST) -n $(NAMESPACE)
	@echo "Aktualisiere Deployment auf $(IMAGE)..."
	kubectl set image deployment/$(DEPLOYMENT_NAME) $(APP_NAME)=$(IMAGE) -n $(NAMESPACE) --record=false 2>/dev/null || \
		kubectl rollout restart deployment/$(DEPLOYMENT_NAME) -n $(NAMESPACE)
	@echo "Warte bis der Pod bereit ist..."
	kubectl rollout status deployment/$(DEPLOYMENT_NAME) -n $(NAMESPACE)

## reload: Baut ohne Cache neu, forciert den Pod-Neustart und startet die Port-Forwards frisch
reload:
	@echo "Erzwinge Rebuild ohne Cache..."
	docker build --no-cache -t $(IMAGE) .
	$(MAKE) deploy
	$(MAKE) port-forward

## port-forward: Startet Forwards im Hintergrund (nutzt Services statt statischer Pod-Namen)
port-forward: stop-forward
	@echo "Starte Port-Forwards..."
	@kubectl port-forward -n $(NAMESPACE) svc/httpserver-svc $(LOCAL_PORT):$(SVC_PORT) > /dev/null 2>&1 & echo $$! >> $(PID_FILE)
	@kubectl port-forward -n $(MON_NAMESPACE) svc/vmsingle-vm-stack-victoria-metrics-k8s-stack $(VM_PORT):$(VM_PORT) > /dev/null 2>&1 & echo $$! >> $(PID_FILE)
	@kubectl port-forward -n $(MON_NAMESPACE) svc/vm-stack-grafana $(GRAFANA_PORT):$(GRAFANA_PORT) > /dev/null 2>&1 & echo $$! >> $(PID_FILE)
	@sleep 1
	@echo "Bereit!"
	@echo "👉 Service: http://localhost:$(LOCAL_PORT)"
	@echo "👉 VMUI:    http://localhost:$(VM_PORT)/vmui"
	@echo "👉 Grafana: http://localhost:$(GRAFANA_PORT)"

## stop-forward: Beendet laufende Hintergrundprozesse sauber
stop-forward:
	@if [ -f $(PID_FILE) ]; then \
		echo "Beende bestehende Port-Forwards..."; \
		while read -r pid; do \
			kill -9 $$pid > /dev/null 2>&1 || true; \
		done < $(PID_FILE); \
		rm -f $(PID_FILE); \
	fi