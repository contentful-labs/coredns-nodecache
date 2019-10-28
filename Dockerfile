FROM golang:1.12-stretch AS builder

RUN apt update && apt upgrade -y && apt install iptables -y

RUN git clone --single-branch --branch v1.6.4 https://github.com/coredns/coredns.git /coredns

WORKDIR /coredns

RUN make gen
RUN make

RUN mkdir -p plugin/nodecache
RUN echo 'nodecache:nodecache' >> /coredns/plugin.cfg

COPY *.go /coredns/plugin/nodecache/
RUN make
RUN chmod 0755 /coredns/coredns

FROM alpine:latest
RUN apk add iptables

COPY --from=builder /coredns/coredns /
COPY Corefile /

EXPOSE 5300

ENTRYPOINT ["/coredns"]
