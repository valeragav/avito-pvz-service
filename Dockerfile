FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o avito-pvz-service ./cmd/app

FROM alpine:3.18 AS runtime

WORKDIR /app

COPY --from=build /app/avito-pvz-service .

RUN apk add --no-cache tzdata ca-certificates

ENV TZ=Europe/Moscow

USER nobody:nobody

EXPOSE 8080 8081 9091 3000

CMD ["./avito-pvz-service"]