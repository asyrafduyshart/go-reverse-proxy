FROM golang:alpine AS builder

COPY . /go/src
WORKDIR /go/src

RUN ls

RUN go get ./
RUN go build -o app

FROM alpine:latest  

WORKDIR /root/

COPY --from=builder /go/src/app app
RUN ls
RUN chmod 0755 ./app

CMD ./app start