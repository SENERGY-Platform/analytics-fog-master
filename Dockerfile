FROM alpine:3.20 AS builder

ARG VERSION=dev

RUN apk add --no-cache --update go gcc g++

COPY . /go/src/app
WORKDIR /go/src/app

ENV GO111MODULE=on

RUN CGO_ENABLED=1 go build -o app cmd/main.go

RUN git log -1 --oneline > version.txt

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/app/app .
COPY --from=builder /go/src/app/version.txt .
COPY --from=builder /go/src/app/migrations migrations

EXPOSE 8080

ENTRYPOINT ["./app"]
