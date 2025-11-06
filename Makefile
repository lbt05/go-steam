CPUS ?= $(shell (nproc --all || sysctl -n hw.ncpu) 2>/dev/null || echo 1)
MAKEFLAGS += --jobs=$(CPUS)
PROTO_FILES := $(shell find protobufs -mindepth 2 -maxdepth 2 -type f -name "*.proto" -path "**/steam/**")

.PHONY: help
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: format
format: ## Format files
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.3.0 fmt ./...

.PHONY: lint
lint: ## Lint files
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.3.0 run ./...

.PHONY: fakes
fakes: ## Generate fakes for unit testing
	@go install go.uber.org/mock/mockgen@latest
	@go generate ./...

.PHONY: build
build: ## Build the application
	@go build ./...

.PHONY: test
test: ## Run tests
	@mkdir -p coverage
	@go run gotest.tools/gotestsum@latest -- \
		-race -count=1 -covermode=atomic \
		-coverprofile=coverage/coverage.cov \
		./...
	@go run github.com/axw/gocov/gocov@latest convert coverage/coverage.cov | go run github.com/AlekSi/gocov-xml@latest > coverage/coverage.xml

.PHONY: bench
bench: ## Run benchmarks
	@go test -bench=. -benchmem ./...

.PHONY: language
language: ## Generate language files from protobufs
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install ./cmd/protoc-gen-enums
	@rm -rf language
	@$(MAKE) $(PROTO_FILES)

define compile_proto
	@mkdir -p language/$(shell basename $(dir $(1)))
	@-protoc \
		--proto_path=$(dir $(1)) \
		--experimental_allow_proto3_optional \
		$(foreach proto,$(wildcard $(dir $(1))*.proto),--go_opt=M$(shell basename $(proto))=./$(shell basename $(dir $(1)))) \
		--go_out=language \
		$(foreach proto,$(wildcard $(dir $(1))*.proto),--enums_opt=M$(shell basename $(proto))=./$(shell basename $(dir $(1)))) \
		--enums_out=language \
		$(1)
endef
define compile_proto_target
.PHONY: $(1)
$(1):
	$$(call compile_proto,$(1))
endef
$(foreach proto,$(PROTO_FILES),$(eval $(call compile_proto_target,$(proto))))
