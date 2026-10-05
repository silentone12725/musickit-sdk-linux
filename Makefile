# Top-level orchestration. Each component also builds on its own:
#   drm/      make -C drm libdrm_client.so
#   sdk/      cd sdk && go test ./...
#   server/   cd server && go build ./cmd/musickit-engine   (stub DRM; use `make engine` for real DRM)

DIST ?= dist

.PHONY: all drm engine test vet fmt clean
all: engine

drm:
	$(MAKE) -C drm libdrm_client.so

# Real engine: CGO + native_backend tag, linked against drm/libdrm_client.so.
engine: drm
	scripts/build-engine.sh $(DIST)

test:
	cd sdk && go test ./...
	cd server && go test ./...
	cd examples && go vet ./...

vet:
	cd sdk && go vet ./...
	cd server && go vet ./...

fmt:
	gofmt -l sdk server examples

clean:
	$(MAKE) -C drm clean
	rm -rf $(DIST)
