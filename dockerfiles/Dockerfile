FROM golang:alpine as build_base

RUN apk add --no-cache git
RUN apk add build-base

# Set the Current Working Directory inside the container
WORKDIR /app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN go build -mod=mod -o ./sso .

FROM alpine:latest

WORKDIR /app

COPY --from=build_base /app/setup /app/setup

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Run the binary program produced by `go install`
ENTRYPOINT ["./sso"]