FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .

ARG CMD_PATH

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./${CMD_PATH}

FROM alpine:latest

RUN apk add --no-cache ffmpeg

WORKDIR /app

COPY --from=build /app/app .

COPY --from=build /app/migrations ./migrations

ENTRYPOINT ["/app/app"]