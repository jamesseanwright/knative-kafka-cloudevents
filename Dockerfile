FROM golang:1.20.1 AS build
ARG entrypoint
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ENV CGO_ENABLED=0
RUN go build -o bin/main ./cmd/${entrypoint}

FROM alpine:3.17.2
COPY --from=build /app/bin/main .
ENTRYPOINT ["./main"]
