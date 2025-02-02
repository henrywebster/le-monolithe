ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .


FROM debian:bookworm

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /run-app /usr/local/bin/
# TODO: Add template files to reasonable directory
COPY --from=builder /usr/src/app/template/ /usr/local/bin/template/
# TODO: Add data.db to reasonable directory
COPY --from=builder /usr/src/app/data.db /usr/local/bin/data.db
CMD ["run-app"]
