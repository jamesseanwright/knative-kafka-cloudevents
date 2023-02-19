FROM golang:1.20.1 AS build
ARG entrypoint
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ENV CGO_ENABLED=0
RUN go build -o bin/main ./cmd/${entrypoint}

FROM gcr.io/distroless/static-debian11
COPY --from=build /app/bin/main .
CMD ["./main"]
