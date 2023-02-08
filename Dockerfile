FROM golang:1.20-alpine3.17 as builder

WORKDIR /

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build main.go

FROM gcr.io/distroless/static-debian11
WORKDIR /api

USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /main /api

EXPOSE 80

ENTRYPOINT ["/api/main"]
