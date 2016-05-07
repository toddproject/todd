#SHELL := /bin/bash

all: install_deps compile

clean:
	rm -f $(GOPATH)/bin/todd-server
	rm -f $(GOPATH)/bin/todd
	rm -f $(GOPATH)/bin/todd-agent

build: compile
	docker build -t mierdin/todd -f Dockerfile .

install:
	go install ./cmd/...

fmt:
	go fmt github.com/mierdin/todd/...

test: 
	godep go test ./... -cover

install_deps:
	go get github.com/tools/godep
	go get -u github.com/jteeuwen/go-bindata/...

update_deps: install_deps
	godep save ./...

update_assets: install_deps
	$(GOPATH)/bin/go-bindata -o assets/assets_unpack.go -pkg="assets" -prefix="agent" agent/testing/testlets/... agent/facts/collectors/...

start: install
	start-containers.sh 3 /etc/todd/server-int.cfg /etc/todd/agent-int.cfg
