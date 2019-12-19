FROM golang:latest
WORKDIR /capi-kpack-watcher

COPY capi-kpack-watcher/ .
RUN cd /capi-kpack-watcher && go build

ENTRYPOINT ["/capi-kpack-watcher/capi_kpack_watcher"]
