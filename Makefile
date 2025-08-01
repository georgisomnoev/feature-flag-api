.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: test
test:
	ginkgo run -tags -race ./...

.PHONY: test-unit
test-unit:
	ginkgo run -tags -race --label-filter '!/./' ./...

.PHONY: test-integration
test-integration:
	ginkgo run -tags -race --label-filter '/./' ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: generate
generate:
	go generate ./...

.PHONY: run-app
run-app:
	docker compose down -v --remove-orphans 
	docker compose up -d --build featureflagsapi featureflagsdb migratedb

.PHONY: run-app-with-otel
run-app-with-otel:
	OTEL_COLLECTOR_HOST="otel-collector:4317" \
	docker compose down -v --remove-orphans 
	docker compose up -d --build

CERTS_DIR ?= certs

.PHONY: generate-certs-and-keys
generate-certs-and-keys: server-certs jwt-keys

.PHONY: server-certs
server-certs:
	@echo "Generating self-signed server certificates"
	@rm -rf $(CERTS_DIR)/server
	@mkdir -p $(CERTS_DIR)/server
	@openssl req -x509 -newkey rsa:2048 \
		-keyout $(CERTS_DIR)/server/server.key \
		-out $(CERTS_DIR)/server/server.crt \
		-days 365 -sha256 -nodes \
		-subj "/CN=localhost"

.PHONY: jwt-keys
jwt-keys:
	@echo "Generating RSA private and public keys for JWT"
	@rm -rf $(CERTS_DIR)/jwt_keys
	@mkdir -p $(CERTS_DIR)/jwt_keys
	@openssl genpkey -algorithm RSA \
		-out $(CERTS_DIR)/jwt_keys/private.pem \
		-pkeyopt rsa_keygen_bits:2048
	openssl rsa -in $(CERTS_DIR)/jwt_keys/private.pem -pubout -out $(CERTS_DIR)/jwt_keys/public.pem