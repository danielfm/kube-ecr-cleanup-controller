GO=CGO_ENABLED=0 go
TAG=0.1.0
BIN=kube-ecr-cleanup-controller
IMAGE=307424997672.dkr.ecr.us-east-1.amazonaws.com/$(BIN)

.PHONY: build
build:
	$(GO) build -a --ldflags '-w -extldflags "-static"' -tags netgo -o bin/$(BIN) ./cmd

.PHONY: image
image: build
	docker build -t $(IMAGE):$(TAG) .

.PHONY: push
push: image
	docker push $(IMAGE):$(TAG)

.PHONY: push-latest
push-latest: image
	docker tag $(IMAGE):$(TAG) $(IMAGE):latest
	docker push $(IMAGE):latest

.PHONY: clean
clean:
	rm -Rf bin/ cover*

.PHONY: test
test:
	./test default

.PHONY: cover
cover:
	./test with-cover
