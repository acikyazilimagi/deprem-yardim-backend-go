FROM golang:1.20-alpine as builder

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o api cmd/api/main.go

FROM alpine:latest as runner

RUN apk add --no-cache bash tini

COPY --from=builder --chown=nonroot:nonroot /app/api /

EXPOSE 80

ENTRYPOINT ["/sbin/tini", "--"]

CMD ["/api"]
