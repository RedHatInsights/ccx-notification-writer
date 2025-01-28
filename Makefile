SHELL := /bin/bash

.PHONY: default clean benchmarks build build-tests fmt lint vet cyclo ineffassign shellcheck errcheck goconst gosec abcgo style run test test-postgres cover integration_tests rest_api_tests sqlite_db license before_commit bdd_tests help godoc install_docgo install_addlicense

SOURCES:=$(shell find . -name '*.go')
BINARY:=ccx-notification-writer
DOCFILES:=$(addprefix docs/packages/, $(addsuffix .html, $(basename ${SOURCES})))

default: build

clean: ## Run go clean
	@go clean

build: ${BINARY} ## Build binary containing service executable

build-cover:	${SOURCES}  ## Build binary with code coverage detection support
	./build.sh -cover

${BINARY}: ${SOURCES}
	./build.sh

fmt: ## Run go fmt -w for all sources
	@echo "Running go formatting"
	./gofmt.sh

lint: ## Run golint
	@echo "Running go lint"
	./golint.sh

vet: ## Run go vet. Report likely mistakes in source code
	@echo "Running go vet"
	./govet.sh

cyclo: ## Run gocyclo
	@echo "Running gocyclo"
	./gocyclo.sh

ineffassign: ## Run ineffassign checker
	@echo "Running ineffassign checker"
	./ineffassign.sh

shellcheck: ## Run shellcheck
	./shellcheck.sh

errcheck: ## Run errcheck
	@echo "Running errcheck"
	./goerrcheck.sh

goconst: ## Run goconst checker
	@echo "Running goconst checker"
	./goconst.sh ${VERBOSE}

gosec: ## Run gosec checker
	@echo "Running gosec checker"
	./gosec.sh ${VERBOSE}

abcgo: ## Run ABC metrics checker
	@echo "Run ABC metrics checker"
	./abcgo.sh ${VERBOSE}

style: fmt vet lint cyclo shellcheck errcheck goconst gosec ineffassign abcgo ## Run all the formatting related commands (fmt, vet, lint, cyclo) + check shell scripts

run: ${BINARY} ## Build the project and executes the binary
	./$^

test: ${BINARY} ## Run the unit tests
	./unit-tests.sh

build-test: ## Build native binary with unit tests and benchmarks
	go test -c

profiler: ${BINARY} ## Run the unit tests with profiler enabled
	./profile.sh

benchmark.txt:	benchmark

benchmark: ${BINARY} ## Run benchmarks
	go test -bench=. -run=^$ | tee benchmark.txt

benchmark.csv:	benchmark.txt ## Export benchmark results into CSV
	awk '/Benchmark/{count ++; gsub(/BenchmarkTest/,""); printf("%d,%s,%s,%s\n",count,$$1,$$2,$$3)}' $< > $@

cover: test ## Generate HTML pages with code coverage
	@go tool cover -html=coverage.out

coverage: ## Display code coverage on terminal
	@go tool cover -func=coverage.out

license: install_addlicense
	addlicense -c "Red Hat, Inc" -l "apache" -v ./

bdd_tests: ## Run BDD tests
	@echo "Run BDD tests"
	pushd bdd_tests/ && ./run_tests.sh && popd

before_commit: style test test-postgres integration_tests license ## Checks done before commit
	./check_coverage.sh

help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z._-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

function_list: ${BINARY} ## List all functions in generated binary file
	go tool objdump ${BINARY} | grep ^TEXT | sed "s/^TEXT\s//g"

docs/packages/%.html: %.go
	mkdir -p $(dir $@)
	docgo -outdir $(dir $@) $^
	addlicense -c "Red Hat, Inc" -l "apache" -v $@

godoc: export GO111MODULE=off
godoc: install_docgo install_addlicense ${DOCFILES} docs/sources.md

docs/sources.md: docs/sources.tmpl.md ${DOCFILES}
	./gen_sources_md.sh

install_docgo: export GO111MODULE=off
install_docgo:
	[[ `command -v docgo` ]] || go get -u github.com/dhconnelly/docgo

install_addlicense: export GO111MODULE=off
install_addlicense:
	[[ `command -v addlicense` ]] || GO111MODULE=off go get -u github.com/google/addlicense
