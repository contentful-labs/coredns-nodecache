FROM golang:1.26.2 AS builder

RUN apt update && apt upgrade -y && apt install iptables -y

RUN git clone --single-branch --branch v1.14.3 https://github.com/coredns/coredns.git /coredns

WORKDIR /coredns

RUN go get -u && go mod tidy && make gen
RUN make

RUN mkdir -p plugin/nodecache
RUN echo 'nodecache:nodecache' >> /coredns/plugin.cfg

COPY *.go /coredns/plugin/nodecache/
RUN make
RUN chmod 0755 /coredns/coredns

FROM alpine:3.24.1
RUN apk add --no-cache iptables

COPY --from=builder /coredns/coredns /
COPY Corefile /

EXPOSE 5300

ENTRYPOINT ["/coredns"]
