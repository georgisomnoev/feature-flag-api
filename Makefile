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

# Set OTEL_COLLECTOR_ENABLED to true in the .env file if you wish to start collecting metrics/traces.
.PHONY: run-app-with-obsv
run-app-with-obsv:
	docker compose down -v --remove-orphans 
	docker compose up -d --build

CERTS_DIR ?= certs
.PHONY: jwt-keys
jwt-keys:
	@echo "Generating RSA private and public keys for JWT"
	@rm -rf $(CERTS_DIR)/jwt_keys
	@mkdir -p $(CERTS_DIR)/jwt_keys
	@openssl genpkey -algorithm RSA \
		-out $(CERTS_DIR)/jwt_keys/private.pem \
		-pkeyopt rsa_keygen_bits:2048
	openssl rsa -in $(CERTS_DIR)/jwt_keys/private.pem -pubout -out $(CERTS_DIR)/jwt_keys/public.pem