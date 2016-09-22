SHELL=/bin/bash

all: compile

clean:
	rm -f $(GOPATH)/bin/todd-server
	rm -f $(GOPATH)/bin/todd
	rm -f $(GOPATH)/bin/todd-agent

build:
	docker build -t mierdin/todd -f Dockerfile .

compile:

	# Installing testlets
	./scripts/gettestlets.sh

	# Installing ToDD
	go install ./cmd/...

fmt:
	go fmt github.com/mierdin/todd/...

test: 
	go test ./... -cover

update_deps:
	go get -u github.com/tools/godep
	godep save ./...

update_assets:
	go get -u github.com/jteeuwen/go-bindata/...
	go-bindata -o assets/assets_unpack.go -pkg="assets" -prefix="agent" agent/testing/bashtestlets/... agent/facts/collectors/...

start: compile
	start-containers.sh 3 /etc/todd/server-int.cfg /etc/todd/agent-int.cfg

install:

	# Set testlet capabilities
	./scripts/set-testlet-capabilities.sh

	# Copy configs if etc and /etc/todd aren't linked
	if ! [ "etc" -ef "/etc/todd" ]; then mkdir -p /etc/todd && cp -f ./etc/{agent,server}.cfg /etc/todd/; fi
	mkdir -p /opt/todd/{agent,server}/assets/{factcollectors,testlets}
	chmod -R 777 /opt/todd

	# If on Linux, enable ping testlet functionality (DEPRECATED in favor of granting socket capabilities on testlets)
	# sysctl -w net.ipv4.ping_group_range="0 0" || echo "Unable to set kernel parameters to allow ping. Some testlets may not work."
