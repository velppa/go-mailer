FROM alpine

MAINTAINER Pavel Popov keybase.io/pavelpopov

RUN apk add --no-cache ca-certificates bash coreutils nodejs

# installing mjml
RUN npm install -g mjml

ADD ./build /app

ENV PATH=/app:$PATH

ENV MAILER_CONFIG=/config.toml

CMD mailer -config $MAILER_CONFIG
