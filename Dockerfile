ARG GOLANG_VERSION=1.10-alpine
FROM golang:${GOLANG_VERSION} as builder
LABEL maintainer "pgillich ta gmail.com"

# Based on:
# https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
# https://medium.com/@pierreprinetti/the-go-dockerfile-d5d43af9ee3c
# https://www.cloudreach.com/blog/containerize-this-golang-dockerfiles/

ARG REPO_NAME="pgillich/prometheus_text-to-remote_write/"
ARG BIN_PATH="/prometheus_text-to-remote_write"

COPY . $GOPATH/src/github.com/${REPO_NAME}
WORKDIR $GOPATH/src/github.com/${REPO_NAME}
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' \
    -o ${BIN_PATH} .

# Making minimal image (only one binary)

FROM scratch

ARG BIN_PATH="/prometheus_text-to-remote_write"

ARG RECEIVE_PORT="9099"

COPY --from=builder ${BIN_PATH} "/prometheus_text-to-remote_write"
ENTRYPOINT ["/prometheus_text-to-remote_write", "service"]

EXPOSE ${RECEIVE_PORT}
