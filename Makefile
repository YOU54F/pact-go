include make/config.mk

TEST?=./...
.DEFAULT_GOAL := ci
DOCKER_HOST_HTTP?="http://host.docker.internal"
PACT_CLI="docker run --rm -v ${PWD}:${PWD} -e PACT_BROKER_BASE_URL=$(DOCKER_HOST_HTTP) -e PACT_BROKER_USERNAME -e PACT_BROKER_PASSWORD pactfoundation/pact-cli"

ifeq ($(OS),Windows_NT)
	EXE=.exe
	SKIP_AVRO=1
endif
ci:: docker deps clean bin test pact
ci_unit:: deps clean bin test
ci_pact:: docker pact
ci_unit_no_docker:: deps clean bin test
ci_pact_no_docker:: pact

# Run the ci target from a developer machine with the environment variables
# set as if it was on Travis CI.
# Use this for quick feedback when playing around with your workflows.
fake_ci:
	@CI=true \
	APP_SHA=`git rev-parse --short HEAD`+`date +%s` \
	APP_BRANCH=`git rev-parse --abbrev-ref HEAD` \
	make ci

# same as above, but just for pact
fake_pact:
	@CI=true \
	APP_SHA=`git rev-parse --short HEAD`+`date +%s` \
	APP_BRANCH=`git rev-parse --abbrev-ref HEAD` \
	make pact

docker:
	@echo "--- 🛠 Starting docker"
	docker-compose up -d

bin:
	go build -o build/pact-go

clean:
	mkdir -p ./examples/pacts
	rm -rf build output dist examples/pacts

deps: download_plugins
	@echo "--- 🐿  Fetching build dependencies "
	cd /tmp; \
	go install github.com/mitchellh/gox@latest; \
	cd -

# avro plugin requires apk add bash and openjdk17-jre
# go plugin linux-aarch64 version works on musl, requires musl named artifact
# csv plugin requires a musl version creating
# protobuf plugin requires apk add protobuf-dev protoc
download_plugins:
	@echo "--- 🐿  Installing plugins"; \
	./scripts/install-cli.sh
	if [ $${SKIP_AVRO:-0} -ne 1 ]; then \
		$$HOME/.pact/bin/pact-plugin-cli$(EXE) -y install https://github.com/austek/pact-avro-plugin/releases/tag/v0.0.5; \
	fi
cli:
	@if [ ! -d pact/bin ]; then\
		echo "--- 🐿 Installing Pact CLI dependencies"; \
		curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash -x; \
	fi

install: bin
	echo "--- 🐿 Installing Pact FFI dependencies"
	./build/pact-go	 -l DEBUG install --libDir /tmp

pact:
	@echo "--- 🔨 Running Pact examples"
	go test -v -count=1 -tags=consumer github.com/pact-foundation/pact-go/v2/examples/...
	go test -v -count=1 -timeout=30s -tags=provider github.com/pact-foundation/pact-go/v2/examples/...

publish:
	@echo "-- 📃 Publishing pacts"
	@"${PACT_CLI}" publish ${PWD}/examples/pacts --consumer-app-version ${APP_SHA} --tag ${APP_BRANCH} --tag prod

release:
	echo "--- 🚀 Releasing it"
	"$(CURDIR)/scripts/release.sh"

test: deps install
	@echo "--- ✅ Running tests"
	@if [ -f coverage.txt ]; then rm coverage.txt; fi;
	@echo "mode: count" > coverage.txt
	@for d in $$(go list ./... | grep -v vendor | grep -v examples); \
		do \
			go test -count=1 -v -coverprofile=profile.out -covermode=atomic $$d; \
			if [ $$? != 0 ]; then \
				export FAILURE=1; \
			fi; \
			if [ -f profile.out ]; then \
					cat profile.out | tail -n +2 >> coverage.txt; \
					rm profile.out; \
			fi; \
	done; \
	if [ $${FAILURE:-0} -eq 1 ]; then \
		exit 1; \
	fi;
	go tool cover -func coverage.txt


testrace:
	go test -race $(TEST) $(TESTARGS)

updatedeps:
	go get -d -v -p 2 ./...

.PHONY: install bin default dev test pact updatedeps clean release
