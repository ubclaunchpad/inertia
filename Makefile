.PHONY: test test-verbose test-profile test-env clean docker bootstrap

PACKAGES = `go list ./... | grep -v vendor/`

all: inertia

inertia:
	go build

test:
	go test $(PACKAGES) --cover

test-verbose:
	go test $(PACKAGES) -v --cover

test-env:
	docker build -t sshvps -f ./test_env/Dockerfile.sshvps ./test_env
	docker run --rm -d -p 22:22 -p 8081:8081 --name testvps --privileged sshvps
	bash ./test_env/info.sh

clean: inertia
	rm -f inertia

docker:
	docker build -t ubclaunchpad/inertia .
	docker push ubclaunchpad/inertia

bootstrap:
	go-bindata -o client/bootstrap.go -pkg client client/bootstrap/...
