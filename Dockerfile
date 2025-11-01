FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /out/api     ./main.go && \
    go build -o /out/seeder  ./cmd/seeder/main.go
FROM alpine:3.20

COPY --from=build /main /app/main
COPY --from=build /seeder /app/seeder



WORKDIR /app


EXPOSE 8080
EXPOSE 8181


ENTRYPOINT ["/app/main"]
