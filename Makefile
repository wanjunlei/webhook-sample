# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

IMG ?= leiwanjun/webhook-sample:latest

all: docker-build

# Build notification-adapter binary
webhook-sample:
	go build -o webhook-sample cmd/main.go

# Build the docker image
docker-build:
	docker buildx build --platform linux/amd64,linux/arm64 --push -f Dockerfile -t ${IMG} .

