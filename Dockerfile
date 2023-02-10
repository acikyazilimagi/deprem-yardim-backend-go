FROM golang:1.20-alpine as builder

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -o api cmd/api/main.go

FROM gcr.io/distroless/static-debian11 as runner

COPY --from=builder --chown=nonroot:nonroot /app/api /

EXPOSE 80

ENTRYPOINT ["/api"]
