FROM golang:1.23

WORKDIR /cmd

COPY go.mod .

RUN go mod download
COPY cmd/app/main.go .

COPY . .

RUN go build -o main ./cmd/app/main.go

EXPOSE 8080

CMD ["./main"]