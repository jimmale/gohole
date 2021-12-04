FROM golang:1 AS builder

ADD ./ /go/src/github.com/jimmale/gohole
WORKDIR /go/src/github.com/jimmale/gohole

ENV CGO_ENABLED=0
RUN go build .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/github.com/jimmale/gohole/gohole /gohole
CMD ["/gohole"]
