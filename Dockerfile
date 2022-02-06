FROM ruby:3.1.0

ENV LANG=C.UTF-8
RUN apt-get update && apt-get install -qq -y build-essential libpq-dev postgresql-client imagemagick --fix-missing --no-install-recommends

RUN wget https://go.dev/dl/go1.17.6.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf  go1.17.6.linux-amd64.tar.gz

ENV app /app
RUN mkdir -p $app
CMD ./main

COPY Gemfile* /tmp/
WORKDIR /tmp
RUN bundle install -j8

WORKDIR $app
ADD . $app
RUN /usr/local/go/bin/go build -o main *.go
