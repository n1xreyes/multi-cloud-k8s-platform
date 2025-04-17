ARG POSTGRES_VERSION=13-alpine
FROM postgres:${POSTGRES_VERSION}
LABEL authors="Angelo Reyes"

# hadolint ignore=DL3018
RUN apk add --no-cache go git bash make

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/bin" && chmod -R 777 "$GOPATH"
RUN mkdir /app

RUN echo "Installing migrate..." && \
    go install -v -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.2

COPY migrations/postgres /app/migrations/postgres
COPY scripts/migrate.sh /app/scripts/migrate.sh
RUN chmod +x /app/scripts/migrate.sh

WORKDIR /app

EXPOSE 5432