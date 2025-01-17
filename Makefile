COVERAGE_FILE ?= coverage.out

TARGET ?= chirpy

.PHONY: build
build: lint 
	@echo "Building ${TARGET}..."
	@mkdir -p .bin
	@go build -o .bin/${TARGET} ./cmd/${TARGET}

.PHONY: run
run: build
	@./.bin/${TARGET}

.PHONY: test
test:
	@go test --race -count=1 ./...

.PHONY: coverage
coverage:
	@go test -coverpkg='github.com/mashfeii/chirpy/...' --race -count=1 -coverprofile='$(COVERAGE_FILE)' ./...
	@go tool cover -html='$(COVERAGE_FILE)' -o coverage.html && xdg-open coverage.html&
	@go tool cover -func='$(COVERAGE_FILE)' | grep ^total | tr -s '\t'

.PHONY: lint
lint:
	@golangci-lint run

.PHONY: clean
clean:
	@rm -rf .bin
