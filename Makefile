TAG = `git describe --tags`
SSH_PORT = 69
VPS_VERSION = latest
VPS_OS = ubuntu
RELEASE = test
CLI_VERSION_VAR = main.Version
PROJECTNAME=inertia

.PHONY: help
help: Makefile
	@echo " Choose a command run in "$(PROJECTNAME)":\n"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'

## all: install production dependencies and build the Inertia CLI to project directory
all: prod-deps cli

## deps: install all dependencies
.PHONY: deps
deps: prod-deps dev-deps docker-deps

## lint: run static analysis for entire project
.PHONY: lint
lint: SHELL:=/bin/bash
lint:
	go vet ./...
	go test -run xxxx ./...
	diff -u <(echo -n) <(gofmt -d -s `find . -type f -name '*.go' -not -path "./vendor/*"`)
	diff -u <(echo -n) <(go run golang.org/x/lint/golint `go list ./... | grep -v /vendor/`)

## clean: remove testenv, binaries, build directories, caches, and more
.PHONY: clean
clean: testenv-clean
	go clean -testcache
	rm -f ./inertia
	rm -f ./inertiad
	rm -f ./inertia-*
	rm -rf ./docs_build
	find . \
		-type f \
		-name inertia.\* \
		-not -path "*.static*" \
		-not -path "*docs*" \
		-exec rm {} \;

##    ____________
##  * CLI / DAEMON
##    ‾‾‾‾‾‾‾‾‾‾‾‾

## cli: build the inertia CLI binary into project directory
.PHONY: cli
cli:
	go build -ldflags "-X $(CLI_VERSION_VAR)=$(RELEASE)"

## install: install inertia CLI to $GOPATH
.PHONY: install
install:
	go install -ldflags "-X $(CLI_VERSION_VAR)=$(RELEASE)"

## daemon: build the daemon image and save it to ./images
.PHONY: daemon
daemon:
	mkdir -p ./images
	rm -f ./images/inertia-daemon-image
	docker build --build-arg INERTIA_VERSION=$(TAG) \
		-t ubclaunchpad/inertia:test .
	docker save -o ./images/inertia-daemon-image ubclaunchpad/inertia:test
	docker rmi ubclaunchpad/inertia:test

## gen: rewrite all generated code (mocks, scripts, etc.)
.PHONY: gen
gen: scripts mocks

##    _______
##  * TESTING
##    ‾‾‾‾‾‾‾

## testenv: set up full test environment
.PHONY: testenv
testenv: docker-deps testenv-clean
	# run nginx container for testing
	docker run --name testcontainer -d nginx

	# start vps container
	docker build -f ./test/vps/$(VPS_OS).dockerfile \
		-t $(VPS_OS)vps \
		--build-arg VERSION=$(VPS_VERSION) \
		./test
	bash ./test/start_vps.sh $(SSH_PORT) $(VPS_OS)vps

## testdaemon: create test daemon and scp the image to the test VPS for use as version "test"
.PHONY: testdaemon
testdaemon: daemon testdaemon-scp

## test: run unit test suite
.PHONY: test
test:
	go test ./... -short -ldflags "-X $(CLI_VERSION_VAR)=test" --cover

## test-all: build test daemon, set up testenv, and run unit and integration tests 
.PHONY: test-all
test-all:
	make testenv VPS_OS=$(VPS_OS) VPS_VERSION=$(VPS_VERSION)
	make testdaemon
	go test ./... -ldflags "-X $(CLI_VERSION_VAR)=test" --cover

## test-integration: build test daemon, set up testenv, and run integration tests only
.PHONY: test-integration
test-integration:
	make testenv VPS_OS=$(VPS_OS) VPS_VERSION=$(VPS_VERSION)
	make testdaemon
	go test ./... -v -run 'Integration' -ldflags "-X $(CLI_VERSION_VAR)=test" --cover

## test-integration-fast: run integration tests only, but without setting up testenv
.PHONY: test-integration-fast
test-integration-fast:
	make testdaemon
	go test ./... -v -run 'Integration' -ldflags "-X $(CLI_VERSION_VAR)=test" --cover

##    _____________
##  * DOCUMENTATION
##    ‾‾‾‾‾‾‾‾‾‾‾‾‾

DOCS_DIR=docs

## docs: build all documentation, used for latest release documentation
.PHONY: docs
docs: docs-usage docs-cli docs-api

## docs-tip: build tip documentation (docs/tip), used for the bleeding edge documentation
.PHONY: docs-tip
docs-tip:
	make docs DOCS_DIR=docs/tip

## docs-usage: set up doc builder and build usage guide website
.PHONY: docs
docs-usage:
	sh .scripts/build_docs.sh $(DOCS_DIR)

## docs-cli: build CLI reference pages
.PHONY: docs-cli
docs-cli: docgen
	@echo [INFO] Generating CLI documentation
	@./inertia-docgen -o $(DOCS_DIR)/cli

## docs-api: build API reference from Swagger definitions in /docs_src/api
.PHONY: docs-api
docs-api: 
	@echo [INFO] Generating API documentation
	@redoc-cli bundle ./docs_src/api/swagger.yml -o $(DOCS_DIR)/api/index.html

## run-docs-usage: run doc server from ./docs_src for the usage guide website only
.PHONY: run-docs-usage
run-docs-usage:
	( cd docs_build/slate ; bundle exec middleman server --verbose )

## run-docs-api: run doc server from ./docs_src for the API reference website only
.PHONY: run-docs-api
run-docs-api:
	redoc-cli serve ./docs_src/api/swagger.yml -w

##    _______
##  * HELPERS
##    ‾‾‾‾‾‾‾

## prod-deps: install only production dependencies
.PHONY: prod-deps
prod-deps:
	go mod download

## dev-deps: install only development dependencies and tools
.PHONY: dev-deps
dev-deps:
	npm install -g redoc-cli

## docker-deps: download required docker containers
.PHONY: docker-deps
docker-deps:
	bash test/docker_deps.sh

## mocks: generate Go mocks
.PHONY: mocks
mocks:
	go run github.com/maxbrunsfeld/counterfeiter/v6 -o ./client/runner/mocks/session.go \
		./client/runner/ssh.go SSHSession
	go run github.com/maxbrunsfeld/counterfeiter/v6 -o ./daemon/inertiad/project/mocks/deployer.go \
		./daemon/inertiad/project/deployment.go Deployer
	go run github.com/maxbrunsfeld/counterfeiter/v6 -o ./daemon/inertiad/build/mocks/builder.go \
		./daemon/inertiad/build/builder.go ContainerBuilder
	go run github.com/maxbrunsfeld/counterfeiter/v6 -o ./daemon/inertiad/notify/mocks/notify.go \
		./daemon/inertiad/notify/notifier.go Notifier

## scripts: recompile script assets
.PHONY: scripts
scripts:
	go run github.com/UnnoTed/fileb0x b0x.yml

## testdaemon-scp: copy test daemon image from ./images to test VPS
.PHONY: testdaemon-scp
testdaemon-scp:
	chmod 400 ./test/keys/id_rsa
	scp -i ./test/keys/id_rsa \
		-o StrictHostKeyChecking=no \
		-o UserKnownHostsFile=/dev/null \
		-P $(SSH_PORT) \
		./images/inertia-daemon-image \
		root@0.0.0.0:/daemon-image

## testenv-clean: stop and shut down the test environment
.PHONY: testenv-clean
testenv-clean:
	docker stop testvps testcontainer || true && docker rm testvps testcontainer || true

##    _______________
##  * RELEASE SCRIPTS
##    ‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾

## install-tagged: install Inertia with git tag as release version
.PHONY: install-tagged
install-tagged:
	go install -ldflags "-X $(CLI_VERSION_VAR)=$(TAG)"

## daemon-release: build the daemon and push it to the UBC Launch Pad Docker Hub
.PHONY: daemon-release
daemon-release:
	docker build --build-arg INERTIA_VERSION=$(RELEASE) \
		-t ubclaunchpad/inertia:$(RELEASE) .
	docker push ubclaunchpad/inertia:$(RELEASE)

## cli-release: cross-compile Inertia CLI binaries for distribution
.PHONY: cli-release
cli-release:
	bash .scripts/release.sh

##    ____________
##  * EXPERIMENTAL
##    ‾‾‾‾‾‾‾‾‾‾‾‾

## contrib: install everything in 'contrib'
.PHONY: contrib
contrib:
	go install  -ldflags "-X $(CLI_VERSION_VAR)=$(TAG)" ./contrib/...

## docgen: build the docgen tool into project directory
.PHONY: docgen
docgen:
	go build -ldflags "-X $(CLI_VERSION_VAR)=$(TAG)" ./contrib/inertia-docgen

## localdaemon: run a test daemon locally, without testenv
.PHONY: localdaemon
localdaemon:
	bash ./test/start_local_daemon.sh
