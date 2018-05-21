TAG = `git describe --tags`
SSH_PORT = 22
VPS_VERSION = latest
VPS_OS = ubuntu
RELEASE = test

all: prod-deps inertia

# List all commands
.PHONY: ls
ls:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | xargs

# Sets up production dependencies
.PHONY: prod-deps
prod-deps:
	dep ensure
	make web-deps

# Sets up test dependencies
.PHONY: dev-deps
dev-deps:
	go get -u github.com/jteeuwen/go-bindata/...
	bash test/deps.sh	

# Install Inertia with release version
.PHONY: inertia
inertia:
	go install -ldflags "-X main.Version=$(RELEASE)"

# Install Inertia with git tag as release version
.PHONY: inertia-tagged
inertia-tagged:
	go install -ldflags "-X main.Version=$(TAG)"

# Remove Inertia binaries
.PHONY: clean
clean:
	rm -f ./inertia
	find . -type f -name inertia.\* -exec rm {} \;

.PHONY: lint
lint:
	PATH=$(PATH):./bin bash -c './bin/gometalinter --vendor --deadline=60s ./...'
	(cd ./daemon/web; npm run lint)

# Run unit test suite
.PHONY: test
test:
	go test ./... -short -ldflags "-X main.Version=test" --cover

# Run unit test suite verbosely
.PHONY: test-v
test-v:
	go test ./... -short -ldflags "-X main.Version=test" -v --cover

# Run unit and integration tests - creates fresh test VPS and test daemon beforehand
# Also attempts to run linter
.PHONY: test-all
test-all:
	make lint
	make testenv VPS_OS=$(VPS_OS) VPS_VERSION=$(VPS_VERSION)
	make testdaemon
	go clean -testcache
	go test ./... -ldflags "-X main.Version=test" --cover

# Run integration tests verbosely - creates fresh test VPS and test daemon beforehand
.PHONY: test-integration
test-integration:
	go clean -testcache
	make testenv VPS_OS=$(VPS_OS) VPS_VERSION=$(VPS_VERSION)
	make testdaemon
	go test ./... -v -run 'Integration' -ldflags "-X main.Version=test" --cover

# Run integration tests verbosely without recreating test VPS
.PHONY: test-integration-fast
test-integration-fast:
	go clean -testcache
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

# Create test daemon and scp the image to the test VPS for use.
# Requires Inertia version to be "test"
.PHONY: testdaemon
testdaemon:
	rm -f ./inertia-daemon-image
	docker build --build-arg INERTIA_VERSION=$(TAG) \
		-t ubclaunchpad/inertia:test .
	docker save -o ./inertia-daemon-image ubclaunchpad/inertia:test
	docker rmi ubclaunchpad/inertia:test
	chmod 400 ./test/keys/id_rsa
	scp -i ./test/keys/id_rsa \
		-o StrictHostKeyChecking=no \
		-o UserKnownHostsFile=/dev/null \
		-P $(SSH_PORT) \
		./inertia-daemon-image \
		root@0.0.0.0:/daemon-image
	rm -f ./inertia-daemon-image

# Creates a daemon release and pushes it to Docker Hub repository.
# Requires access to the UBC Launch Pad Docker Hub.
.PHONY: daemon
daemon:
	docker build --build-arg INERTIA_VERSION=$(RELEASE) \
		-t ubclaunchpad/inertia:$(RELEASE) .
	docker push ubclaunchpad/inertia:$(RELEASE)

# Recompiles assets. Use whenever a script in client/bootstrap is
# modified.
.PHONY: bootstrap
bootstrap:
	go-bindata -o client/bootstrap.go -pkg client client/bootstrap/...

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
