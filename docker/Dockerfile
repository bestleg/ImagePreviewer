# syntax=docker/dockerfile:1

FROM golang:1.17-alpine AS build

WORKDIR app/image-previewer

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o /image-previewer ./cmd/main.go

EXPOSE 8081

CMD [ "/image-previewer" ]
