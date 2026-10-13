FROM golang:1.27.1 AS builder

RUN apt update && apt install -y iptables

RUN git clone --single-branch --branch v1.14.7 https://github.com/coredns/coredns.git /coredns

WORKDIR /coredns

RUN go get -u google.golang.org/grpc@v1.83.2 && \
    go get -u golang.org/x/crypto@v0.56.0 && \
    go get -u golang.org/x/mod@v0.40.0 && \
    go mod tidy && \
    make gen && make # update pkgs manually

RUN mkdir -p plugin/nodecache
RUN echo 'nodecache:nodecache' >> /coredns/plugin.cfg

COPY *.go /coredns/plugin/nodecache/
RUN make
RUN chmod 0755 /coredns/coredns

FROM alpine:3.24.2

RUN apk add --no-cache iptables "zlib>=1.3.2-r1"

COPY --from=builder /coredns/coredns /
COPY Corefile /

EXPOSE 5300

ENTRYPOINT ["/coredns"]
