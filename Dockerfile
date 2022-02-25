FROM golang:1.17
COPY . /go/src/github.com/tjmerritt/udp-replicate
WORKDIR /go/src/github.com/tjmerritt/udp-replicate
RUN cd cmd/read-test; go build
RUN cd cmd/write-test; go build
RUN cd cmd/udp-replicate; go build
ENTRYPOINT ["/cmd/udp-replicate/upd-replicate"]
