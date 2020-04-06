FROM golang:1.13 as builder
WORKDIR /capi-kpack-watcher

# Copy the go source
COPY go.mod .
COPY go.sum .

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Build
COPY / .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM cloudfoundry/run:tiny
WORKDIR /
COPY --from=builder /capi-kpack-watcher/capi_kpack_watcher .
USER nonroot:nonroot

ENTRYPOINT ["/capi_kpack_watcher"]
