FROM golang:1.16-buster AS builder

RUN apt update && apt upgrade -y && apt install iptables -y

RUN git clone --single-branch --branch v1.8.4 https://github.com/coredns/coredns.git /coredns

WORKDIR /coredns

RUN make gen
RUN make

RUN mkdir -p plugin/nodecache
RUN echo 'nodecache:nodecache' >> /coredns/plugin.cfg

COPY *.go /coredns/plugin/nodecache/
RUN go get github.com/coreos/go-iptables@f901d6c2a4f2a4df092b98c33366dfba1f93d7a0 github.com/vishvananda/netlink@f049be6f391489d3f374498fe0c8df8449258372
RUN make
RUN chmod 0755 /coredns/coredns

FROM alpine:3.13
RUN apk add iptables

COPY --from=builder /coredns/coredns /
COPY Corefile /

EXPOSE 5300

ENTRYPOINT ["/coredns"]
