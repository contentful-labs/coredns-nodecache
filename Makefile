build:
	docker buildx build --platform linux/amd64 -t contentful-labs/coredns-nodecache-amd64 .
	docker buildx build --platform linux/arm64 -t contentful-labs/coredns-nodecache-arm64 .
	docker tag contentful-labs/coredns-nodecache-amd64 contentful-labs/coredns-nodecache

run: build
	docker run --cap-add=NET_ADMIN --cap-add=NET_RAW --privileged -P contentful-labs/coredns-nodecache

test:
	docker run -ti -v $$PWD:/go/src/github.com/contentful-labs/coredns-nodecache \
	-w /go/src/github.com/contentful-labs/coredns-nodecache/ golang:1.18-buster go test -v -mod=vendor ./...

lint:
	docker run -ti -v $$PWD:/go/src/github.com/contentful-labs/coredns-nodecache \
	-w /go/src/github.com/contentful-labs/coredns-nodecache/ golangci/golangci-lint:v1.46.2 golangci-lint run
