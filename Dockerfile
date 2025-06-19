FROM golang:1.23.0-bullseye AS gobuild
WORKDIR /go/src/github.com/chaunceyjiang/single-process-oom
COPY . .
RUN make build

FROM ubuntu:20.04
WORKDIR /single-process-oom

COPY --from=gobuild /go/src/github.com/chaunceyjiang/single-process-oom/single-process-oom /bin/single-process-oom

ENTRYPOINT ["/bin/single-process-oom"]
