# API and Comms Re-Write Design

## Problems with Current Approach

- No centralized models for objects passed via API
- No versioning
- Comms is a separate thing - we can and should make one API for everything
- No authentication
- Inconsistent UX on the CLI and API

## New Solution

- All objects defined in protobuf, centrally. One place (API) to see definitions for all ToDD primitives. Getting rid of "objects" nomenclature.
- API for interacting with ToDD as well as agent communication
- Use grpc authentication for securit

## TODO

- [ ] See if you can solve the asset distribution problem with the new api. Shouldn't need to do interface detection on the server.
- [ ] How to version grpc/protobuf APIs? Need to create models AND the code implementation in an easily versionable way
- [ ] Install grpc and protobuf dependencies automatically for CI, and add these steps to docs too.
- [ ] API docs as well as architectural doc update re: the new purpose of API
- [ ] How to write tests for all this?
- [ ] Need to also handle live groups somehow. Should this be part of this service or should it be a field that shows up in Agents? Assuming it should be the latter, need to maintain group membership in the local agents store somehow, instead of just an in-memory dict
- [ ] Need to fix deps....maybe not in this branch, but the current way of adding deps is just not good enough. Or fix etcd bullshit.
- [ ] client needs to be built to easily be used as a lib not just a cli
- [ ] Replace existing API calls first before moving on to rethinking comms - but you do need to put some thought into how that side will work in case it impacts rest of API design. Need to actually create a design doc after some playing around with code before dedicating a lot of time to this.
- [ ] To create more advanced assertion functionality, you're gonna have to require testlets to provide a data schema probably.

## Links

https://grpc.io/docs/guides/auth.html
https://medium.com/@shijuvar/building-high-performance-apis-in-go-using-grpc-and-protocol-buffers-2eda5b80771b
https://cloud.google.com/apis/design/versioning
https://cloud.google.com/apis/design/compatibility
https://grpc.io/docs/tutorials/basic/go.html#generating-client-and-server-code
https://coreos.com/blog/grpc-protobufs-swagger.html
https://github.com/pseudomuto/protoc-gen-doc


## Installing Protobuf

https://gist.github.com/sofyanhadia/37787e5ed098c97919b8c593f0ec44d8
go get -u google.golang.org/grpc
go get -u github.com/grpc/grpc-go
go get -u github.com/golang/protobuf/protoc-gen-go

godep go install google.golang.org/grpc
godep go install github.com/grpc/grpc-go
godep go install github.com/golang/protobuf/protoc-gen-go