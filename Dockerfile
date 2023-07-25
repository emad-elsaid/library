FROM golang:1.17.7-alpine

ENV LANG=C.UTF-8
RUN apk update && apk add --no-cache postgresql-client

ENV app /app
RUN mkdir -p $app
CMD ./main

WORKDIR $app
ADD . $app
RUN go build -o main *.go
