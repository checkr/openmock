FROM golang:1.12 as builder
WORKDIR /go/src/github.com/checkr/openmock
ADD . .
RUN make build

FROM alpine:3.6
WORKDIR /go/src/github.com/checkr/openmock
RUN apk add --no-cache ca-certificates libc6-compat
COPY --from=builder /go/src/github.com/checkr/openmock/om ./om
ENV OPENMOCK_HTTP_HOST=0.0.0.0
ENV OPENMOCK_TEMPLATES_DIR=/data/templates
CMD ./om
