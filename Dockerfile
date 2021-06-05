FROM ruby:3.0.1

ENV LANG=C.UTF-8
RUN apt-get update && apt-get install -qq -y build-essential libpq-dev postgresql-client --fix-missing --no-install-recommends

ENV app /app
RUN mkdir -p $app
CMD bundle exec ./main

COPY Gemfile* /tmp/
WORKDIR /tmp
RUN bundle install -j8

WORKDIR $app
ADD . $app
RUN bundle outdated --strict
