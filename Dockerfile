FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN apk add --no-cache bash protobuf protobuf-dev git make vim

RUN go build -o avito-pvz-service ./cmd/app

ENV PATH="/go/bin:$PATH"

EXPOSE 8080

CMD ["./avito-pvz-service"]
