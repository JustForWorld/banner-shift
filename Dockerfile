FROM golang:1.22.2 AS modules

WORKDIR /modules

COPY go.mod go.sum ./

RUN go mod download

FROM golang:1.22.2 AS builder

COPY --from=modules /go/pkg /go/pkg

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o /bin/banner-shift ./cmd/banner-shift

FROM alpine:latest

COPY --from=builder /app/config/local.yaml /config/local.yaml

COPY --from=builder /bin/banner-shift /banner-shift

EXPOSE 8080

CMD ["/banner-shift", "--config", "/config/local.yaml"]