FROM golang:1.20-alpine3.17 as builder

WORKDIR /

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build main.go

FROM gcr.io/distroless/static-debian11
WORKDIR /api

COPY --from=builder /main /api

USER nonroot:nonroot

EXPOSE 8000

ENTRYPOINT ["/api"]
