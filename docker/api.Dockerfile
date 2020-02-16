FROM golang:latest 

ENV GOPATH=/go

RUN mkdir -p /go/src/api
COPY ./cmd/api /go/src/api
COPY config.json.template /go/src/api/config.json

WORKDIR /go/src/api
RUN go get ./...
RUN go build . 
CMD ["./api", "./config.json"]
