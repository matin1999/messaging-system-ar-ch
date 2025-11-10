FROM golang:1.25.0-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /app/main /app/main

EXPOSE 8282
EXPOSE 8181

ENTRYPOINT ["/app/main"]
