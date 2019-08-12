build:
	docker build -t contentful/coredns-nodecache .

run: build
	docker run --cap-add=NET_ADMIN --cap-add=NET_RAW --privileged -P contentful/coredns-nodecache

test: build
	docker run -ti -w="/coredns/plugin/nodecache" contentful/coredns-nodecache go test -mod=vendor -v ./...
