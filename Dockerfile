FROM golang:1.26 AS build

WORKDIR /app
COPY . .
RUN go clean --modcache
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build ./main.go

FROM alpine:latest

RUN apk add --no-cache curl tzdata

WORKDIR /root
COPY --from=build /app/main .
COPY --from=build /app/.env .

EXPOSE 8888
CMD ["./main"]
