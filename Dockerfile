FROM golang:1.23.0 AS gobuild
WORKDIR /go/src/github.com/chaunceyjiang/single-process-oom
COPY . .
RUN make  build-cmd

FROM ubuntu:22.04
WORKDIR /single-process-oom

COPY --from=gobuild /go/src/github.com/chaunceyjiang/single-process-oom/single-process-oom /bin/single-process-oom

ENTRYPOINT ["/bin/single-process-oom"]
