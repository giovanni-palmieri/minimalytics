FROM golang:1.24.1-alpine3.21

RUN apk add --no-cache gcc musl-dev sqlite-dev

ENV CGO_ENABLED=1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o minim .

EXPOSE 3333

CMD ["./minim", "execserver"]