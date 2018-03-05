.PHONY: test test-verbose test-profile testenv clean daemon testdaemon bootstrap

TAG = `git describe --tags`
PACKAGES = `go list ./... | grep -v vendor/`
SSH_PORT = 22
VPS_VERSION = latest
VPS_OS = ubuntu
RELEASE = test

all: inertia

inertia:
	go install -ldflags "-X main.Version=$(RELEASE)"

inertia-tagged:
	go install -ldflags "-X main.Version=$(TAG)"

test:
	make testenv VPS_OS=$(VPS_OS) VPS_VERSION=$(VPS_VERSION)
	make testdaemon
	go test $(PACKAGES) -ldflags "-X main.Version=$(RELEASE)" --cover

test-verbose:
	make testenv VPS_OS=$(VPS_OS) VPS_VERSION=$(VPS_VERSION)
	make testdaemon	
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

testdaemon:
	rm -f ./inertia-daemon-image
	docker build -t ubclaunchpad/inertia:test .
	docker save -o ./inertia-daemon-image ubclaunchpad/inertia:test
	chmod 700 ./test_env/test_key
	scp -i ./test_env/test_key \
		./inertia-daemon-image \
		root@0.0.0.0:/daemon-image
	rm -f ./inertia-daemon-image

daemon:
	docker build -t ubclaunchpad/inertia:$(RELEASE) .
	docker push ubclaunchpad/inertia:$(RELEASE)

bootstrap:
	go-bindata -o client/bootstrap.go -pkg client client/bootstrap/...
