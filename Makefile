export GOPROXY=https://goproxy.cn,direct
export GO111MODULE=on

ARCH=$(shell uname -p)
ifeq ($(ARCH), x86_64)
        ARCH=amd64
else ifeq ($(ARCH), aarch64)
        ARCH=arm64
endif

OBJS=kube-relay test

all: dep ${OBJS}

kube-relay:
	CGO_ENABLED=0 go build -mod vendor -gcflags "-N -l" -o $@ ./cmd/kube-relay

test:
	go build -mod vendor -gcflags "-N -l" -o $@ ./cmd/test

clean:
	rm -fr ${OBJS}

-include .deps
dep:
	echo '${OBJS}: \\' > .deps
	find . -path ./vendor -prune -o -name '*.go' -print | awk '{print $$0 " \\"}' >> .deps
	echo "" >> .deps

docker:
	docker build -f deploy/docker/Dockerfile  . -t ir0cn/kube-relay:$(ARCH)
