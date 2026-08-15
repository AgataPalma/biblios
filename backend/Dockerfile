FROM golang:1.25.8-alpine
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6
RUN go install github.com/air-verse/air@v1.64.5

COPY . .
CMD ["air"]
