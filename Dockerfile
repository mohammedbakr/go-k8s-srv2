FROM golang:alpine AS builder
WORKDIR /go/src/github.com/k8-proxy/go-k8s-srv2
COPY . .
RUN cd cmd \
    && env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o  go-k8s-srv2 .

FROM alpine
COPY --from=builder /go/src/github.com/k8-proxy/go-k8s-srv2/cmd/go-k8s-srv2 /bin/go-k8s-srv2

RUN apk update && apk add ca-certificates

ENTRYPOINT ["/bin/go-k8s-srv2"]
