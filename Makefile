.PHONY: test test-verbose test-profile testenv clean docker bootstrap

TAG = `git describe --tags`
PACKAGES = `go list ./... | grep -v vendor/`
SSH_PORT = 22
VPS_VERSION = latest
VPS_OS = ubuntu
RELEASE = canary

all: inertia

inertia:
	go install -ldflags "-X main.Version=$(RELEASE)"

inertia-tagged:
	go install -ldflags "-X main.Version=$(TAG)"

test:
	make testenv VPS_OS=$(VPS_OS) VPS_VERSION=$(VPS_VERSION)
	go test $(PACKAGES) -ldflags "-X main.Version=$(RELEASE)" --cover

test-verbose:
	make testenv VPS_OS=$(VPS_OS) VPS_VERSION=$(VPS_VERSION)
	go test $(PACKAGES) -ldflags "-X main.Version=$(RELEASE)" -v --cover

testenv:
	docker stop testvps || true && docker rm testvps || true
	docker build -f ./test_env/Dockerfile.$(VPS_OS) \
		-t $(VPS_OS)vps \
		--build-arg VERSION=$(VPS_VERSION) \
		./test_env
	bash ./test_env/startvps.sh $(SSH_PORT) $(VPS_OS)vps

clean:
	rm -f inertia 
	find . -type f -name inertia.\* -exec rm {} \;

docker:
	docker build -t ubclaunchpad/inertia:$(RELEASE) .
	docker push ubclaunchpad/inertia:$(RELEASE)

bootstrap:
	go-bindata -o client/bootstrap.go -pkg client client/bootstrap/...
