module := $(shell head -n1 go.mod | sed -e 's/module //')
version := $(shell git describe --tags)

.PHONY: cert
cert:
	@openssl req -newkey rsa:2048 \
		-new -nodes -x509 \
		-days 365 \
		-out server.crt \
		-keyout server.key \
		-subj "/C=RU/ST=Moscow/L=Moscow/O=Yandex Practicum/OU=Go Developer/CN=localhost" \
		2> /dev/null

.PHONY: test
test:
	@go test -short -timeout=30s -count=1 -cover ./...

.PHONY: install
install:
	@go install -ldflags "-s -w -X '$(module)/version.Version=$(version)'" ./cmd/gk

.PHONY: compose.up
compose.up:
	@docker-compose -f ./deployments/docker-compose.yml up -d

.PHONY: compose.down
compose.down:
	@docker-compose -f ./deployments/docker-compose.yml down

