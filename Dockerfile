FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/bot ./cmd/bot && CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate

FROM alpine:3.20
RUN adduser -D -h /app appuser
WORKDIR /app
COPY --from=build /out/bot /app/bot
COPY --from=build /out/migrate /app/migrate
COPY migrations /app/migrations
USER appuser
EXPOSE 8080
CMD ["/app/bot"]
