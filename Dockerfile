FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/mattermost-bot

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/mattermost-bot /usr/local/bin/mattermost-bot

WORKDIR /app

ENTRYPOINT [ "/usr/local/bin/mattermost-bot" ]

EXPOSE 8000