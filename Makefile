.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build pacman
	go build -o pacman

.PHONY: run
run: build certs ## Build and run pacman
	USE_MTLS=true \
	TLS_ROOT_CA=$$(cat certs/PacMan_Root_CA.crt) \
	TLS_SERVER_CERT=$$(cat certs/localhost.crt) \
	TLS_SERVER_KEY=$$(cat certs/localhost.key) \
		./pacman

.PHONY: test
test: ## Run tests
	go test -cover -race ./...

.PHONY: mocks
mocks: ## Generate mocks for unit tests
	mockgen -package=main -mock_names=registry=RegistryMock \
		-source registry.go -destination=mock_registry.go registry
	mockgen -package=main -mock_names=handler=HandlerMock \
		-source handler.go -destination=mock_handler.go handler
	mockgen -package=main -mock_names=Conn=NetConnMock \
		-destination=mock_net_conn.go net Conn
	mockgen -package=main -mock_names=Listener=NetListenerMock \
		-destination=mock_net_listener.go net Listener

.PHONY: cover
cover: ## Generate test coverage report
	@echo "mode: count" > coverage.out
	@go test -coverprofile coverage.tmp ./...
	@tail -n +2 coverage.tmp >> coverage.out
	@go tool cover -html=coverage.out

certs: ## Generate mTLS certs
	# create CA
	certstrap --depot-path certs init --common-name "PacMan Root CA"
	# create server cert request
	certstrap --depot-path certs request-cert --domain localhost
	# create client cert request
	certstrap --depot-path certs request-cert --cn pacman_client
	# sign server and client cert requests
	certstrap --depot-path certs sign --CA "PacMan Root CA" localhost
	certstrap --depot-path certs sign --CA "PacMan Root CA" pacman_client
	@tree -hrC certs

OPENSSL_CLIENT := openssl s_client -quiet -no_ign_eof -connect localhost:9000 -cert certs/pacman_client.crt -key certs/pacman_client.key

.PHONY: add
add: ## Add a package, usage: make add name='name' deps='dep1 dep2'
	(echo 'AddPackage $(name) $(deps)'; sleep 0.5) | $(OPENSSL_CLIENT)

.PHONY: remove
remove: ## Remove a package, usage: make remove name='name'
	(echo 'RemovePackage $(name)'; sleep 0.5) | $(OPENSSL_CLIENT)

.PHONY: list
list: ## List packages, usage: make list
	(echo 'ListPackages'; sleep 0.5) | $(OPENSSL_CLIENT)

.PHONY: seed
seed: ## Seed pacman with some test data
	@make add name='AAA'
	@make add name='BBB' deps='AAA'
	@make add name='CCC' deps='BBB'
	@make add name='DDD' deps='AAA BBB'
	@make add name='EEE' deps='DDD'
