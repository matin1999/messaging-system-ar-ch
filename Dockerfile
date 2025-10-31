FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN  go build   /main ./main

FROM alpine:3.20


COPY --from=build /mina /app/main


WORKDIR /app


EXPOSE 8080
EXPOSE 8181


ENTRYPOINT ["/app/main"]
