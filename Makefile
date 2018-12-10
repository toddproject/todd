SHELL=/bin/bash

all: compile

clean:
	rm -f $(GOPATH)/bin/todd-server
	rm -f $(GOPATH)/bin/todd
	rm -f $(GOPATH)/bin/todd-agent

build:
	rm -f api/v1/generated/*
	docker build -t toddproject/todd -f Dockerfile .

compile:

	@echo "Generating protobuf code..."

	@rm -f pkg/ui/data/swagger/datafile.go

	@rm -f /tmp/datafile.go
	@rm -f cmd/syringed/buildinfo.go

	@rm -rf api/exp/generated/ && mkdir -p api/exp/generated/
	@mkdir -p api/exp/swagger/

	@protoc -I api/exp/definitions/ -I. \
	-I api/exp/definitions/ \
	  api/exp/definitions/*.proto \
		-I$$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-I$$GOPATH/src/github.com/lyft/protoc-gen-validate \
	--go_out=plugins=grpc:api/exp/generated/ \
    --grpc-gateway_out=logtostderr=true,allow_delete_body=true:api/exp/generated/ \
    --validate_out=lang=go:api/exp/generated/ \
	--swagger_out=logtostderr=true,allow_delete_body=true:api/exp/swagger/

	@# Adding equivalent YAML tags so we can import lesson definitions into protobuf-created structs
	# @sed -i'.bak' -e 's/\(protobuf.*json\):"\([^,]*\)/\1:"\2,omitempty" yaml:"\l\2/' api/exp/generated/lessondef.pb.go
	# @rm -f api/exp/generated/lessondef.pb.go.bak

	@echo "Generating swagger definitions..."
	@go generate ./api/exp/swagger/
	@scripts/build-ui.sh

	@echo "Generating build info file..."
	@scripts/gen-build-info.sh

	# @echo "Installing testlets..."
	# ./scripts/gettestlets.sh

	@echo "Compiling todd binaries..."

ifeq ($(shell uname), Darwin)
	@go install ./cmd/...
else
	@go install -ldflags "-linkmode external -extldflags -static" ./cmd/...
endif


fmt:
	go fmt github.com/toddproject/todd/...

test: 
	go test ./api/... -cover -coverprofile=c.out
	go tool cover -html=c.out -o coverage.html 

	# scripts/start-containers.sh integration



lint:
	scripts/lint.sh

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
	start-containers.sh 6 /etc/todd/server-int.cfg /etc/todd/agent-int.cfg

install:

	# Set capabilities on testlets
	./scripts/set-testlet-capabilities.sh

	# Copy configs if etc and /etc/todd aren't linked
	if ! [ "etc" -ef "/etc/todd" ]; then mkdir -p /etc/todd && cp -f ./etc/{agent,server}.cfg /etc/todd/; fi
	mkdir -p /opt/todd/{agent,server}/assets/{factcollectors,testlets}
	chmod -R 777 /opt/todd
