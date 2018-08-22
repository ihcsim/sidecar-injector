FROM golang:1.10.2 as builder
MAINTAINER Ivan Sim
WORKDIR /go/src/github.com/ihcsim/sidecar-injector
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o server github.com/ihcsim/sidecar-injector/cmd/server

FROM alpine:latest
MAINTAINER Ivan Sim
ARG BUILD_DATE
ARG VCS_REF
ARG VCS_URL
ARG VERSION
WORKDIR /root/
COPY --from=builder /go/src/github.com/ihcsim/sidecar-injector/server .
ENTRYPOINT ["./server"]
LABEL org.label-schema.name="sidecar-injector" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.build-date=${BUILD_DATE} \
      org.label-schema.vcs-ref=${VCS_REF} \
      org.label-schema.vcs-url=${VCS_URL} \
      org.label-schema.version=${VERSION}
