build:
	docker build -t contentful-labs/coredns-nodecache .

run: build
	docker run --cap-add=NET_ADMIN --cap-add=NET_RAW --privileged -P contentful-labs/coredns-nodecache

test: 
	docker run -ti -e GO111MODULE=on -v $$PWD:/go/src/github.com/contentful-labs/coredns-nodecache \
	-w /go/src/github.com/contentful-labs/coredns-nodecache/ golang:1.13-stretch go test -v -mod=vendor ./...
