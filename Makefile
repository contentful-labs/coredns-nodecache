build:
	docker build -t contentful/coredns-nodecache .

run: build
	docker run --cap-add=NET_ADMIN --cap-add=NET_RAW --privileged -P contentful/coredns-nodecache

test: 
	go test -v
