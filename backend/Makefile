# Include the .env file if it exists
ifneq (,$(wildcard .env))
    include .env
    export
endif

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## run/api: run the application
.PHONY: run/api
run/api:
	go run ./cmd/api/

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## run/front: run the application with coors rules relaxed
.PHONY: run/front
run/front:
	go run ./cmd/api/ -cors-trusted-origins="http://localhost:3000 http://localhost:8081 https://localhost:3000 https://localhost:8081"

## client/cert: generate a self-signed cert for the dummy client's https server
.PHONY: client/cert
client/cert:
	openssl req -x509 -newkey rsa:2048 -nodes -days 365 -subj '/CN=localhost' \
		-keyout dummy_client/key.pem -out dummy_client/cert.pem

## client/run: serve the dummy client at https://localhost:4443
.PHONY: client/run
client/run:
	cd dummy_client && python3 server.py

## db/connect: connect to db
.PHONY: db/connect
db/connect:
	psql -h localhost -U postgres -d music

## db/up: run db up migrations
.PHONY: db/up
db/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${DNS_CONN_DEV} up

## db/goto: run migrations to specific version
.PHONY: db/goto
db/goto:
	@echo 'Running goto migration to ${version}...'
	migrate -path ./migrations -database ${DNS_CONN_DEV} goto ${version}

## db/force: run migrations with force after error is fixed
.PHONY: db/force
db/force:
	@echo 'Forcing migration to version ${version}...'
	migrate -path ./migrations -database ${DNS_CONN_DEV} force ${version}

## migration/create: create new migration files
.PHONY: migration/create
migration/create:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit: tidy
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## tidy: tidy and verify module dependencies
.PHONY: tidy
tidy:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags='-s' -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api

# ==================================================================================== #
# PRODUCTION
# ==================================================================================== #

production_host_alias := music-tutor

## production/connect: Connect to the production server
.PHONY: production/connect
production/connect:
	ssh $(production_host_alias)

## production/deploy/api: Deploy the API to production
.PHONY: production/deploy/api
production/deploy/api:
	rsync -P ./bin/linux_amd64/api $(production_host_alias):~
	rsync -rP --delete ./migrations $(production_host_alias):~
	rsync -P ./remote/production/api.service $(production_host_alias):~
	rsync -P ./remote/production/Caddyfile ${production_host_alias}:~
	ssh -t $(production_host_alias) '\
		migrate -path ~/migrations -database $$MUSIC_DB_DSN up \
		&& sudo mv ~/api.service /etc/systemd/system/ \
		&& sudo systemctl enable api \
		&& sudo systemctl restart api \
		&& sudo mv ~/Caddyfile /etc/caddy/ \
		&& sudo systemctl reload caddy \
	'

## production/deploy/books: Deploy the raw text files for embeddings
.PHONY: production/deploy/books
production/deploy/books:
	rsync -rP ./books/ $(production_host_alias):~/books/

## production/deploy/env: copy .env file to server. This will override the existing .env
.PHONY: production/deploy/env
production/deploy/env: confirm
	rsync -P .env $(production_host_alias):~
