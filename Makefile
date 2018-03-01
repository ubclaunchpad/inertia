.PHONY: test test-verbose test-profile testenv clean docker bootstrap

BUILD = `git describe --tags`
PACKAGES = `go list ./... | grep -v vendor/`
SSH_PORT = 22
VERSION = latest
VPS_OS = ubuntu

all: inertia

inertia:
	go install -ldflags "-X main.Version=$(BUILD)"

test:
	make testenv VPS_OS=$(VPS_OS) VERSION=$(VERSION)
	go test $(PACKAGES) --cover

test-verbose:
	make testenv VPS_OS=$(VPS_OS) VERSION=$(VERSION)	
	go test $(PACKAGES) -v --cover

testenv:
	docker stop testvps || true && docker rm testvps || true
	docker build -f ./test_env/Dockerfile.$(VPS_OS) \
		-t $(VPS_OS)vps \
		--build-arg VERSION=$(VERSION) \
		./test_env
	bash ./test_env/startvps.sh $(SSH_PORT) $(VPS_OS)vps

clean: inertia
	rm -f inertia

docker:
	docker build -t ubclaunchpad/inertia .
	docker push ubclaunchpad/inertia

bootstrap:
	go-bindata -o client/bootstrap.go -pkg client client/bootstrap/...
