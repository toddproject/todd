SHELL=/bin/bash

all: compile

clean:
	rm -f $(GOPATH)/bin/todd-server
	rm -f $(GOPATH)/bin/todd
	rm -f $(GOPATH)/bin/todd-agent

build:
	docker build -t mierdin/todd -f Dockerfile .

compile:
	# TODO(mierdin): The current gettestlets.sh script downloads the testlets from
	# Github, meaning a developer would already have to have changes pushed to
	# those repos' master.
	#
	# In a follow-up patch, a script will be provided that allows for rapid development
	# (uses local repositories instead of pulling from GH). Something like devstack.

	# Installing testlets
	./scripts/gettestlets.sh

	# Installing ToDD
	go install ./cmd/...

install: configureenv

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

	# This mode is just to get a demo of ToDD running within the VM quickly.
	# It made sense to re-use the configurations for integration testing, so
	# that's why "server-int.cfg" and "agent-int.cfg" are being used here.
	start-containers.sh 3 /etc/todd/server-int.cfg /etc/todd/agent-int.cfg

configureenv:
	# Copy configs if etc and /etc/todd aren't linked
	if ! [ "etc" -ef "/etc/todd" ]; then mkdir -p /etc/todd && cp -f ./etc/{agent,server}.cfg /etc/todd/; fi
	mkdir -p /opt/todd/{agent,server}/assets/{factcollectors,testlets}
	chmod -R 777 /opt/todd

	# If on Linux, enable ping testlet functionality
	sysctl -w net.ipv4.ping_group_range="0 0" || echo "Unable to set kernel parameters to allow ping. Some testlets may not work."
