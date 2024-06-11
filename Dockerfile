FROM golang:1.22.2

RUN go install github.com/air-verse/air@latest

CMD air