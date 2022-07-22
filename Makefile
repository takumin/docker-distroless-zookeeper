#
# Environment Variables
#

IMAGE_NAME           ?= distroless-zookeeper
IMAGE_REPOSITORY     ?= takumi/$(IMAGE_NAME)
SERVICE_IMAGE_TAG    ?= $(IMAGE_REPOSITORY):latest
ARTIFACT_IMAGE_TAG   ?= $(IMAGE_REPOSITORY):artifact
ENTRYPOINT_IMAGE_TAG ?= $(IMAGE_REPOSITORY):entrypoint

#
# Docker Build Variables
#

export DOCKER_BUILDKIT=1

BUILD_ARGS ?=
BUILD_ARGS += --build-arg BUILDKIT_INLINE_CACHE=1

ifneq (x${no_proxy}x,xx)
BUILD_ARGS += --build-arg no_proxy=${no_proxy}
endif
ifneq (x${NO_PROXY}x,xx)
BUILD_ARGS += --build-arg NO_PROXY=${NO_PROXY}
endif

ifneq (x${ftp_proxy}x,xx)
BUILD_ARGS += --build-arg ftp_proxy=${ftp_proxy}
endif
ifneq (x${FTP_PROXY}x,xx)
BUILD_ARGS += --build-arg FTP_PROXY=${FTP_PROXY}
endif

ifneq (x${http_proxy}x,xx)
BUILD_ARGS += --build-arg http_proxy=${http_proxy}
endif
ifneq (x${HTTP_PROXY}x,xx)
BUILD_ARGS += --build-arg HTTP_PROXY=${HTTP_PROXY}
endif

ifneq (x${https_proxy}x,xx)
BUILD_ARGS += --build-arg https_proxy=${https_proxy}
endif
ifneq (x${HTTPS_PROXY}x,xx)
BUILD_ARGS += --build-arg HTTPS_PROXY=${HTTPS_PROXY}
endif

CACHE_FROM ?= --cache-from $(ARTIFACT_IMAGE_TAG),$(ENTRYPOINT_IMAGE_TAG)

#
# Default Rules
#

.PHONY: default
default: lint build up

#
# Build Rules
#

.PHONY: lint
lint:
	@hadolint Dockerfile

#
# Build Rules
#

.PHONY: build
build:
	@docker build --target service -t $(SERVICE_IMAGE_TAG) $(BUILD_ARGS) $(CACHE_FROM) .
	@docker build --target artifact -t $(ARTIFACT_IMAGE_TAG) $(BUILD_ARGS) .
	@docker build --target entrypoint -t $(ENTRYPOINT_IMAGE_TAG) $(BUILD_ARGS) .

#
# Run Rules
#

.PHONY: run-service
run-service:
	@docker run --rm -i -t -p 12181:2181 --name zookeeper-service $(SERVICE_IMAGE_TAG)

.PHONY: run-artifact
run-artifact:
	@docker run --rm -i -t -p 12181:2181 --name zookeeper-artifact $(ARTIFACT_IMAGE_TAG)

.PHONY: run-entrypoint
run-entrypoint:
	@docker run --rm -i -t -p 12181:2181 --name zookeeper-entrypoint $(ENTRYPOINT_IMAGE_TAG)

#
# Run Rules
#

.PHONY: up
up: down
	@docker compose up

.PHONY: down
down:
ifneq (x$(shell docker ps -aq)x,xx)
	@docker compose down
endif

#
# Clean Rules
#

.PHONY: clean
clean: down
	@docker system prune -f
	@docker volume prune -f
