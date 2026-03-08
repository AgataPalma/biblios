FROM golang:1.25.8-alpine
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/air-verse/air@latest

COPY . .
CMD ["air"]
