.PHONY: inertia test test-verbose testenv clean daemon testdaemon bootstrap

TAG = `git describe --tags`
PACKAGES = `go list ./... | grep -v vendor/`
SSH_PORT = 22
VPS_VERSION = latest
VPS_OS = ubuntu
RELEASE = canary

all: inertia

# Install Inertia with release version
inertia:
	go install -ldflags "-X main.Version=$(RELEASE)"

# Install Inertia with git tag as release version
inertia-tagged:
	go install -ldflags "-X main.Version=$(TAG)"

# Remove binaries
clean:
	rm -f ./inertia 
	find . -type f -name inertia.\* -exec rm {} \;

# Run test suite - creates test VPS and test daemon beforehand
test:
	make testenv VPS_OS=$(VPS_OS) VPS_VERSION=$(VPS_VERSION)
	make testdaemon
	go test $(PACKAGES) -ldflags "-X main.Version=test" --cover

# Run test suite - creates test VPS and test daemon beforehand
test-verbose:
	make testenv VPS_OS=$(VPS_OS) VPS_VERSION=$(VPS_VERSION)
	make testdaemon	
	go test $(PACKAGES) -ldflags "-X main.Version=test" -v --cover

# Run test suite without recreating test VPS
test-dirty:
	make testdaemon
	go test $(PACKAGES) -ldflags "-X main.Version=test" --cover

# Create test VPS
testenv:
	docker stop testvps || true && docker rm testvps || true
	docker build -f ./test_env/Dockerfile.$(VPS_OS) \
		-t $(VPS_OS)vps \
		--build-arg VERSION=$(VPS_VERSION) \
		./test_env
	bash ./test_env/startvps.sh $(SSH_PORT) $(VPS_OS)vps

# Create test daemon and scp the image to the test VPS for use.
# Requires Inertia version to be "test"
testdaemon:
	rm -f ./inertia-daemon-image
	docker build -t ubclaunchpad/inertia:test .
	docker save -o ./inertia-daemon-image ubclaunchpad/inertia:test
	chmod 400 ./test_env/test_key
	scp -i ./test_env/test_key \
		-o StrictHostKeyChecking=no \
		-o UserKnownHostsFile=/dev/null \
		-P $(SSH_PORT) \
		./inertia-daemon-image \
		root@0.0.0.0:/daemon-image
	rm -f ./inertia-daemon-image

# Creates a daemon release and pushes it to Docker Hub repository.
# Requires access to the UBC Launch Pad Docker Hub.
daemon:
	docker build -t ubclaunchpad/inertia:$(RELEASE) .
	docker push ubclaunchpad/inertia:$(RELEASE)

# Recompiles assets. Use whenever a script in client/bootstrap is
# modified.
bootstrap:
	go-bindata -o client/bootstrap.go -pkg client client/bootstrap/...

# Run local development instance of Inertia web.
web-dev:
	(cd ./daemon/web; npm install; npm start)

# Build and minify Inertia web.
web-build:
	(cd ./daemon/web; npm install --production; npm run build)
