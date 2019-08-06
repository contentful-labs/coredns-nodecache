FROM golang:1.12-stretch

RUN apt update && apt upgrade -y && apt install iptables -y

RUN git clone --single-branch --branch v1.5.2 https://github.com/coredns/coredns.git /coredns

WORKDIR /coredns

RUN make gen
RUN make

RUN mkdir -p plugin/nodecache
RUN echo 'nodecache:nodecache' >> /coredns/plugin.cfg

COPY Corefile /coredns/Corefile
COPY *.go /coredns/plugin/nodecache/
RUN make
RUN chmod 0755 /coredns/coredns

FROM alpine:latest
RUN apk add iptables

COPY --from=0 /coredns/coredns /
COPY --from=0 /coredns/Corefile /

EXPOSE 5300

ENTRYPOINT ["/coredns"]
