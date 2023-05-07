# syntax=docker/dockerfile:1

FROM golang:latest AS build
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./reddit_notify

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/reddit_notify /app/reddit_notify
RUN chmod +x ./reddit_notify
CMD [ "./reddit_notify" ]