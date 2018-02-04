.PHONY: test test-verbose test-profile testenv-ubuntu clean docker bootstrap

PACKAGES = `go list ./... | grep -v vendor/`
UBUNTU_VERSION = 16.04

all: inertia

inertia:
	go build

test:
	make testenv-ubuntu VERSION=$(UBUNTU_VERSION)
	go test $(PACKAGES) --cover

test-verbose:
	make testenv-ubuntu VERSION=$(UBUNTU_VERSION)	
	go test $(PACKAGES) -v --cover

testenv-ubuntu:
	docker build -f ./test_env/Dockerfile.ubuntu \
		-t ubuntuvps \
		--build-arg VERSION=$(UBUNTU_VERSION) \
		./test_env
	docker run --rm -d \
		-p 22:22 -p 8081:8081 \
		--name testvps \
		--privileged \
		ubuntuvps
	bash ./test_env/info.sh

clean: inertia
	rm -f inertia

docker:
	docker build -t ubclaunchpad/inertia .
	docker push ubclaunchpad/inertia

bootstrap:
	go-bindata -o client/bootstrap.go -pkg client client/bootstrap/...
