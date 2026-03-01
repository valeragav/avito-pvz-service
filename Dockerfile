FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o avito-pvz-service ./cmd/app

FROM alpine:3.18 AS runtime

WORKDIR /app

COPY --from=build /app/avito-pvz-service .

RUN apk add --no-cache tzdata ca-certificates

ENV TZ=Europe/Moscow

EXPOSE 8080 8081 9091 3000

CMD ["./avito-pvz-service"]