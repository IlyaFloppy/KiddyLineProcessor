# download modules
FROM golang:1.15-alpine as modules

ADD go.mod go.sum /m/
RUN cd /m && go mod download

# build app
FROM golang:1.15-alpine as builder
COPY --from=modules /go/pkg /go/pkg

RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

ENV USER=kiddy
ENV UID=10001

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"
WORKDIR /src/
COPY . .

RUN CGO_ENABLED=0 go build -o /kiddy .

# run app
FROM scratch

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /kiddy /kiddy

USER kiddy:kiddy

ENTRYPOINT ["/kiddy"]