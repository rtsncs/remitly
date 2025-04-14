FROM golang:alpine AS builder

ENV CGO_ENABLED=0
WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /usr/local/bin/app

FROM scratch

WORKDIR /app
COPY --from=builder /usr/local/bin/app ./
EXPOSE 8080

ENTRYPOINT ["./app", "serve"]
