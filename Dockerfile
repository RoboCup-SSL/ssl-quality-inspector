FROM golang:1.25-alpine@sha256:f6751d823c26342f9506c03797d2527668d095b0a15f1862cddb4d927a7a4ced AS build_go
ARG cmd
WORKDIR work
COPY . .
RUN go install ./cmd/${cmd}

# Start fresh from a smaller image
FROM alpine:3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659
ARG cmd
COPY --from=build_go /go/bin/${cmd} /app
USER 1000
ENTRYPOINT ["/app"]
CMD []
