GO=go
GO111MODULE=on
IMAGE_VERSION  ?= $(shell git describe --tags --dirty 2> /dev/null || git rev-parse --short HEAD)
export IMAGE_VERSION
IMAGE_NAME ?= single-process-oom
export IMAGE_NAME
IMAGE_REPOSITORY ?= chaunceyjiang


build:
	$(GO) build -o single-process-oom ./main.go

.PHONY: docker-build
docker-build:
	docker build --build-arg BUILD_TYPE=$(BUILD_TYPE) -t $(IMAGE_REPOSITORY)/$(IMAGE_NAME):$(IMAGE_VERSION) -f Dockerfile .


images: docker-build
	@echo "========== save images =========="
	@mkdir -p $(OUTPUT_DIR)
	@docker tag $(IMAGE_REPOSITORY)/$(IMAGE_NAME):$(IMAGE_VERSION) $(IMAGE_REPOSITORY)/$(IMAGE_NAME):latest
	@docker save -o $(OUTPUT_DIR)/single-process-oom.tar $(IMAGE_REPOSITORY)/$(IMAGE_NAME):latest
	