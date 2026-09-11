FROM golang:1.27.1 AS builder

RUN apt update && apt upgrade -y && apt install iptables -y

RUN git clone --single-branch --branch v1.14.7 https://github.com/coredns/coredns.git /coredns

WORKDIR /coredns

RUN go mod tidy && make gen && make

RUN mkdir -p plugin/nodecache
RUN echo 'nodecache:nodecache' >> /coredns/plugin.cfg

COPY *.go /coredns/plugin/nodecache/
RUN make
RUN chmod 0755 /coredns/coredns

FROM alpine:3.24.1
RUN apk add --no-cache iptables libssl3=3.5.8-r0 libcrypto3=3.5.8-r0

COPY --from=builder /coredns/coredns /
COPY Corefile /

EXPOSE 5300

ENTRYPOINT ["/coredns"]
