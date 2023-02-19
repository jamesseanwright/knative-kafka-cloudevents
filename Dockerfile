FROM golang:1.20.1 AS build
ARG entrypoint
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -o /usr/local/bin/app ./cmd/${entrypoint}

FROM gcr.io/distroless/static-debian11
COPY --from=build /usr/local/bin/app .
CMD ["./app"]
