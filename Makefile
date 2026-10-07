# Top-level orchestration. Each component also builds on its own:
#   drm/      make -C drm libdrm_client.so
#   sdk/      cd sdk && go test ./...
#   server/   cd server && go build ./cmd/musickit-engine   (stub DRM; use `make engine` for real DRM)

DIST ?= dist

.PHONY: all drm engine test vet fmt clean distclean release
all: engine

drm:
	$(MAKE) -C drm libdrm_client.so

# Real engine: CGO + widevine_backend tag, linked against drm/libdrm_client.so.
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

# Intermediates only: compiler objects and test binaries.
clean:
	$(MAKE) -C drm clean

# Everything built, including drm/libdrm_client.so and $(DIST). scripts/clean.sh does the
# same with --dry-run and reporting.
distclean: clean
	rm -rf $(DIST)

# Real engine, then drop the build intermediates.
release: engine clean
