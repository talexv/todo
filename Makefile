override PROJECT_ROOT := $(shell sh -c "git rev-parse --git-dir | xargs dirname | xargs realpath")

test:
	go test -race -cover ./... -args -env ${PROJECT_ROOT}/etc/.env.test

lint:
	golangci-lint run ./...
