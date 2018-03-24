FROM golang:1.9
MAINTAINER Matt Oswalt <matt@keepingitclassless.net> (@mierdin)

LABEL version="0.1"

env PATH /go/bin:$PATH

RUN mkdir /etc/todd

RUN mkdir -p /opt/todd/agent/assets/factcollectors
RUN mkdir -p /opt/todd/server/assets/factcollectors
RUN mkdir -p /opt/todd/agent/assets/testlets
RUN mkdir -p /opt/todd/server/assets/testlets

RUN apt-get update \
 && apt-get install -y vim curl iperf git sqlite3 unzip

RUN curl -OL https://github.com/google/protobuf/releases/download/v3.2.0/protoc-3.2.0-linux-x86_64.zip && unzip protoc-3.2.0-linux-x86_64.zip -d protoc3 && chmod +x protoc3/bin/* && mv protoc3/bin/* /usr/local/bin && mv protoc3/include/* /usr/local/include/

# TODO(mierdin) How to install this vendored? Since we need the binary
RUN go get -u github.com/golang/protobuf/protoc-gen-go

# TODO(mierdin) Fix
RUN go get golang.org/x/net/context
RUN go get google.golang.org/grpc

# Install ToDD
COPY . /go/src/github.com/toddproject/todd

RUN cd /go/src/github.com/toddproject/todd && GO15VENDOREXPERIMENT=1 make && make install

RUN cp /go/src/github.com/toddproject/todd/etc/* /etc/todd


