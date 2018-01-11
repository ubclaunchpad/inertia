.PHONY: test test-verbose test-profile test-race clean docker bootstrap

PACKAGES = `go list ./... | grep -v vendor/`

all: inertia

inertia:
	go build

test:
	go test $(PACKAGES) --cover

test-verbose:
	go test $(PACKAGES) -v --cover

test-race:
	go test $(PACKAGES) -race --cover

clean: inertia
	rm -f inertia
	rm -f cover.html

docker:
	docker build -t ubclaunchpad/inertia .
	docker push ubclaunchpad/inertia

bootstrap:
	go-bindata -o client/bootstrap.go -pkg client client/bootstrap/...
