GO=go
GO111MODULE=on

build:
	$(GO) build -o single-process-oom ./main.go
