FROM golang:1.22.7-alpine

WORKDIR /app

COPY go.mod go.sum credentials.json ./

RUN go mod download

COPY . .

RUN go build -o mura main.go

EXPOSE 8080

# Run the Go app
CMD ["./mura"]
