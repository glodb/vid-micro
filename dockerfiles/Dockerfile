FROM golang:alpine as build_base

RUN apk add --no-cache git
RUN apk add build-base

VOLUME /root/.aws

# Set the Current Working Directory inside the container
WORKDIR /app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN go build -mod=mod -o ./vidmicro .

FROM alpine:latest
RUN apk add ca-certificates
WORKDIR /app

COPY --from=build_base /app/vidmicro /app
COPY --from=build_base /app/config /app/config
COPY --from=build_base /app/com.code.vidmicro/services/authservice/config /app/com.code.vidmicro/services/authservice/config
COPY --from=build_base /app/com.code.vidmicro/services/contentservice/config /app/com.code.vidmicro/services/contentservice/config
COPY --from=build_base /app/com.code.vidmicro/services/titlesservice/config /app/com.code.vidmicro/services/titlesservice/config
COPY --from=build_base /app/com.code.vidmicro/services/sessionscleanerservice/config/ /app/com.code.vidmicro/services/sessionscleanerservice/config/

# Run the binary program produced by `go install`
ENTRYPOINT ["./vidmicro"]