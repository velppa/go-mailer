FROM node:lts-alpine

MAINTAINER Pavel Popov velppa.github.io

RUN apk add --no-cache ca-certificates bash coreutils

# installing mjml
RUN npm install -g mjml

ADD bin/mailer-linux /app/mailer

ENV PATH=/app:$PATH

ENV MAILER_CONFIG=/config.toml

CMD mailer -config $MAILER_CONFIG
