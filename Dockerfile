FROM golang:1.17.7-bullseye

ENV LANG=C.UTF-8
RUN apt-get update && apt-get install -qq -y postgresql-client

ENV app /app
RUN mkdir -p $app
CMD ./main

WORKDIR $app
ADD . $app
RUN go build -o main *.go
