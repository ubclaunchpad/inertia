TAG = `git describe --tags`
SSH_PORT = 69
VPS_VERSION = latest
VPS_OS = ubuntu
RELEASE = test

all: prod-deps cli

# List all commands
.PHONY: ls
ls:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | xargs

# Install all dependencies
.PHONY: deps
deps: prod-deps dev-deps

# Sets up production dependencies
.PHONY: prod-deps
prod-deps:
	dep ensure
	make web-deps

# Sets up test dependencies
.PHONY: dev-deps
dev-deps:
	go get -u github.com/UnnoTed/fileb0x
	bash test/docker_deps.sh
	bash test/lint_deps.sh

# Install Inertia with release version
.PHONY: cli
cli:
	go install -ldflags "-X main.Version=$(RELEASE)"

# Install Inertia with git tag as release version
.PHONY: cli-tagged
cli-tagged:
	go install -ldflags "-X main.Version=$(TAG)"

# Remove Inertia binaries
.PHONY: clean
clean:
	go clean -testcache
	rm -f ./inertia
	find . -type f -name inertia.\* -exec rm {} \;

# Run static analysis
.PHONY: lint
lint:
	PATH=$(PATH):./bin bash -c './bin/gometalinter --vendor --deadline=120s ./...'
	(cd ./daemon/web; npm run lint)

# Run test suite without Docker ops
.PHONY: test
test:
	go test ./... -short -ldflags "-X main.Version=test" --cover

# Run test suite without Docker ops
.PHONY: test-v
test-v:
	go test ./... -short -ldflags "-X main.Version=test" -v --cover

# Run unit and integration tests - creates fresh test VPS and test daemon beforehand
# Also attempts to run linter
.PHONY: test-all
test-all:
	make testenv VPS_OS=$(VPS_OS) VPS_VERSION=$(VPS_VERSION)
	make testdaemon
	go test ./... -ldflags "-X main.Version=test" --cover

# Run integration tests verbosely - creates fresh test VPS and test daemon beforehand
.PHONY: test-integration
test-integration:
	make testenv VPS_OS=$(VPS_OS) VPS_VERSION=$(VPS_VERSION)
	make testdaemon
	go test ./... -v -run 'Integration' -ldflags "-X main.Version=test" --cover

# Run integration tests verbosely without recreating test VPS
.PHONY: test-integration-fast
test-integration-fast:
	make testdaemon
	go test ./... -v -run 'Integration' -ldflags "-X main.Version=test" --cover

# Create test VPS
.PHONY: testenv
testenv:
	docker stop testvps || true && docker rm testvps || true
	docker build -f ./test/vps/Dockerfile.$(VPS_OS) \
		-t $(VPS_OS)vps \
		--build-arg VERSION=$(VPS_VERSION) \
		./test
	bash ./test/start_vps.sh $(SSH_PORT) $(VPS_OS)vps

# Builds test daemon image and saves as inertia-daemon-image
.PHONY: testdaemon-image
testdaemon-image:
	mkdir -p ./images
	rm -f ./images/inertia-daemon-image
	docker build --build-arg INERTIA_VERSION=$(TAG) \
		-t ubclaunchpad/inertia:test .
	docker save -o ./images/inertia-daemon-image ubclaunchpad/inertia:test
	docker rmi ubclaunchpad/inertia:test

# Copies test daemon image to test VPS.
.PHONY: testdaemon-scp
testdaemon-scp:
	chmod 400 ./test/keys/id_rsa
	scp -i ./test/keys/id_rsa \
		-o StrictHostKeyChecking=no \
		-o UserKnownHostsFile=/dev/null \
		-P $(SSH_PORT) \
		./images/inertia-daemon-image \
		root@0.0.0.0:/daemon-image

# Create test daemon and scp the image to the test VPS for use.
# Requires Inertia version to be "test"
.PHONY: testdaemon
testdaemon: testdaemon-image testdaemon-scp

# Run a test daemon locally
.PHONY: localdaemon
localdaemon:
	bash ./test/start_local_daemon.sh

# Creates a daemon release and pushes it to Docker Hub repository.
# Requires access to the UBC Launch Pad Docker Hub.
.PHONY: daemon
daemon:
	docker build --build-arg INERTIA_VERSION=$(RELEASE) \
		-t ubclaunchpad/inertia:$(RELEASE) .
	docker push ubclaunchpad/inertia:$(RELEASE)

# Recompiles assets. Use whenever a script in client/scripts is modified.
.PHONY: scripts
scripts:
	fileb0x b0x.yml

# Install Inertia Web dependencies. Use PACKAGE to install something.
.PHONY: web-deps
web-deps:
	(cd ./daemon/web; npm install $(PACKAGE))

# Run local development instance of Inertia Web.
.PHONY: web-run
web-run:
	(cd ./daemon/web; npm start)

# Build and minify Inertia Web.
.PHONY: web-build
web-build:
	(cd ./daemon/web; npm install --production; npm run build)
