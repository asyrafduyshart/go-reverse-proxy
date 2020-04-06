FROM golang:alpine

ENV APP_NAME go-proxy
ENV PORT 9999

COPY . /go/src/${APP_NAME}
WORKDIR /go/src/${APP_NAME}

RUN go get ./
RUN go build -o ${APP_NAME}

CMD ./${APP_NAME} start

EXPOSE ${PORT}