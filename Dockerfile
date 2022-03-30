FROM golang:1.14-alpine AS build
WORKDIR /go/src/github.com/RoboCup-SSL/ssl-quality-inspector
COPY go.mod go.mod
COPY cmd cmd
COPY pkg pkg
RUN go install ./...

# Start fresh from a smaller image
FROM alpine:3.15
COPY --from=build /go/bin/ssl-quality-inspector /app/ssl-quality-inspector
ENTRYPOINT ["/app/ssl-quality-inspector"]
CMD []
