.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: test
test:
	ginkgo test -race ./...	

.PHONY: run
run:
	go run ./cmd/feature_flags_api/main.go

.PHONY: run-docker
run-docker:
	docker compose down -v --remove-orphans 
	docker compose up --build

CERTS_DIR ?= certs
SERVER_DIR := $(CERTS_DIR)/server

.PHONY: generate-certs
generate-certs: server

.PHONY: server
server:
	@echo "==> Generating self-signed server certificates"
	@rm -rf $(SERVER_DIR)
	@mkdir -p $(SERVER_DIR)
	openssl req -x509 -newkey rsa:2048 \
		-keyout $(SERVER_DIR)/server.key \
		-out $(SERVER_DIR)/server.crt \
		-days 365 -sha256 -nodes \
		-subj "/CN=localhost"