.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build:
	go build -o pacman

.PHONY: run
run: build certs
	USE_MTLS=true \
	TLS_ROOT_CA=$$(cat certs/PacMan_Root_CA.crt) \
	TLS_SERVER_CERT=$$(cat certs/localhost.crt) \
	TLS_SERVER_KEY=$$(cat certs/localhost.key) \
		./pacman

.PHONY: test
test:
	go test ./...

certs:
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
add:
	(echo 'AddPackage $(name) $(deps)'; sleep 0.5) | $(OPENSSL_CLIENT)

.PHONY: remove
remove:
	(echo 'RemovePackage $(name)'; sleep 0.5) | $(OPENSSL_CLIENT)

.PHONY: list
list:
	(echo 'ListPackages'; sleep 0.5) | $(OPENSSL_CLIENT)

.PHONY: seed
seed:
	@make add name='AAA'
	@make add name='BBB' deps='AAA'
	@make add name='CCC' deps='BBB'
	@make add name='DDD' deps='AAA BBB'
	@make add name='EEE' deps='DDD'
