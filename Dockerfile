FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/danielstefank/kleinanzeigen-alert/
COPY . .
RUN mkdir -p /configs
COPY ./configs /configs
RUN go get -d -v

RUN go build -o /go/bin/hello

FROM scratch
COPY --from=builder /go/bin/hello /go/bin/hello
COPY --from=builder /configs /go/bin/configs
ENTRYPOINT ["/go/bin/hello"]