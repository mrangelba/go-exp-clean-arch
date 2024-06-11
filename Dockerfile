FROM golang:1.22.2

RUN go install github.com/cosmtrek/air@latest

CMD air