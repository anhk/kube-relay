export GOPROXY=https://goproxy.cn,direct
export GO111MODULE=on

ARCH=$(shell uname -p)
ifeq ($(ARCH), x86_64)
        ARCH=amd64
else ifeq ($(ARCH), aarch64)
        ARCH=arm64
endif

all: dep kube-relay

kube-relay:
	CGO_ENABLED=0 go build -mod vendor -gcflags "-N -l" -o $@ ./cmd/kube-relay

clean:
	rm -fr kube-relay

-include .deps
dep:
	echo 'kube-relay: \\' > .deps
	find . -path ./vendor -prune -o -name '*.go' -print | awk '{print $$0 " \\"}' >> .deps
	echo "" >> .deps

docker:
	docker build -f deploy/docker/Dockerfile  . -t ir0cn/kube-relay:$(ARCH)
