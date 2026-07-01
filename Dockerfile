FROM golang:1.26-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1 AS build_go
ARG cmd
WORKDIR work
COPY . .
RUN go install ./cmd/${cmd}

# Start fresh from a smaller image
FROM alpine:3@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b
ARG cmd
COPY --from=build_go /go/bin/${cmd} /app
USER 1000
ENTRYPOINT ["/app"]
CMD []
