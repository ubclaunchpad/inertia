.PHONY: test test-verbose test-profile clean docker bootstrap

PACKAGES = `go list ./... | grep -v vendor/`

all: inertia

inertia:
	go build

test:
	go test $(PACKAGES) --cover

test-verbose:
	go test $(PACKAGES) -v --cover

clean: inertia
	rm -f inertia
	rm -f cover.html

docker:
	docker build -t ubclaunchpad/inertia .
	docker push ubclaunchpad/inertia

bootstrap:
	go-bindata -o client/bootstrap.go client/bootstrap/...
