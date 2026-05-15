FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /stockd ./cmd/stockd

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /stockd /usr/local/bin/stockd
EXPOSE 8443
ENTRYPOINT ["stockd"]
