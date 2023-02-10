FROM golang:1.20-alpine3.17 as builder

WORKDIR /

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o api cmd/api/main.go
RUN CGO_ENABLED=0 go build -o consumer cmd/consumer/main.go

FROM gcr.io/distroless/static-debian11

COPY --from=builder --chown=nonroot:nonroot /api /
COPY --from=builder --chown=nonroot:nonroot /consumer /

EXPOSE 80

ENTRYPOINT ["/api"]
