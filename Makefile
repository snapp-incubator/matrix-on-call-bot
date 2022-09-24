export APP=matrix-on-call-bot

export DSN="mysql://on-call:secret@tcp(localhost:33060)/on-call?readTimeout=3s&timeout=30s&parseTime=True"

all: format lint build

run-server:
	go run ./cmd/matrix-on-call-bot server

build:
	go build ./cmd/matrix-on-call-bot

mod:
	go mod download

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o matrix-on-call-bot ./cmd/matrix-on-call-bot

install:
	go install ./cmd/matrix-on-call-bot

############################################################
# Format and Lint
############################################################

check-formatter:
	which goimports || go install golang.org/x/tools/cmd/goimports
	which gofumpt || go install mvdan.cc/gofumpt

format: check-formatter
	find . -type f -name "*.go" -not -path "./vendor/*" | xargs -n 1 -I R goimports -w R
	find . -type f -name "*.go" -not -path "./vendor/*" | xargs -n 1 -I R gofumpt -w R

check-linter:
	which golangci-lint || go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2

lint: check-linter
	golangci-lint run ./...

lint-result: check-linter
	golangci-lint run ./... && echo "Lint Success ✅" || echo "Lint Failure ❌"

############################################################
# Test
############################################################

test:
	go test -v -race -p 1 `go list ./... | grep -v integration`

ci-test:
	go test -v -race -p 1 -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -func coverage.txt

integration-tests:
	go test -v -race -p 1 `go list ./... | grep integration`

############################################################
# Migrations
############################################################

migrate-create:
	migrate create -ext sql -dir ./migrations $(NAME)

migrate-up:
	migrate -verbose -path ./migrations -database $(DSN) up

migrate-down:
	migrate -path ./migrations -database $(DSN) down

migrate-reset:
	migrate -path ./migrations -database $(DSN) drop

migrate-install:
	which migrate || GO111MODULE=off go get -tags 'mysql' -v -u github.com/golang-migrate/migrate/cmd/migrate

############################################################
# Development Environment
############################################################

up:
	docker-compose up

down:
	docker-compose down