build:
	docker build -t contentful/coredns-nodecache .

run: build
	docker run --cap-add=NET_ADMIN --cap-add=NET_RAW --privileged -P contentful/coredns-nodecache

test:
	docker run -it -e GO111MODULE=on -v $$PWD:/go/src/github.com/contentful-labs/coredns-nodecache \
	-w /go/src/github.com/contentful-labs/coredns-nodecache/ golang:1.12-stretch go test -v -mod=vendor ./...