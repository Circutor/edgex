#
# Copyright (c) 2018 Cavium
#
# SPDX-License-Identifier: Apache-2.0
#


.PHONY: build arm clean test run

GO=go
VERSION=$(shell cat ./VERSION)
GOFLAGS=-ldflags "-X github.com/Circutor/edgex.Version=$(VERSION)"

MICROSERVICES=cmd/export-client/export-client cmd/export-distro/export-distro cmd/core-metadata/core-metadata cmd/core-data/core-data cmd/core-command/core-command cmd/support-logging/support-logging cmd/support-notifications/support-notifications cmd/sys-mgmt-agent/sys-mgmt-agent cmd/support-scheduler/support-scheduler cmd/edgex/edgex

.PHONY: $(MICROSERVICES)

build: $(MICROSERVICES)

cmd/core-metadata/core-metadata:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/core-metadata

cmd/core-data/core-data:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/core-data

cmd/core-command/core-command:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/core-command

cmd/export-client/export-client:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/export-client

cmd/export-distro/export-distro:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/export-distro

cmd/support-logging/support-logging:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/support-logging

cmd/support-notifications/support-notifications:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/support-notifications

cmd/sys-mgmt-agent/sys-mgmt-agent:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/sys-mgmt-agent

cmd/support-scheduler/support-scheduler:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/support-scheduler

cmd/edgex/edgex:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/edgex

arm:
	GOOS=linux GOARCH=arm $(GO) build $(GOFLAGS) -o cmd/edgex/edgex ./cmd/edgex

clean:
	rm -f $(MICROSERVICES)

test:
	go test -cover ./...
	go vet ./...

run:
	cd bin && ./edgex-launch.sh

