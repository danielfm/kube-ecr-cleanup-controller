GO=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go
TAG=0.1.10
BIN=kube-ecr-cleanup-controller
IMAGE=danielfm/$(BIN)

.PHONY: deps
deps:
	$(GO) mod tidy

.PHONY: build
build: deps
	$(GO) build -a --ldflags "-X main.VERSION=$(TAG) -w -extldflags '-static'" -tags netgo -o bin/$(BIN) ./cmd/$(BIN)

.PHONY: image
image: build
	podman build -t $(IMAGE):$(TAG) .

.PHONY: push
push: image
	podman push $(IMAGE):$(TAG)

.PHONY: push-latest
push-latest: image
	podman tag $(IMAGE):$(TAG) $(IMAGE):latest
	podman push $(IMAGE):latest

.PHONY: clean
clean:
	rm -Rf bin/ cover*

.PHONY: test
test:
	./test default

.PHONY: cover
cover:
	./test with-cover
