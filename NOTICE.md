# Third-party code and credits

This project is MIT licensed (see `LICENSE`). It includes or depends on the following; each retains its own licence.

| Component | Where | Notes |
|---|---|---|
| Reference host-native DRM wrapper (WorldObservationLog/zhaarey wrapper lineage, with local lease-recovery and library-mode changes) | `drm/native/` (`main.c`, `main.cpp`, `hybris_stubs.c`, `drm_lib.c`, `cmdline.*`, `import.h`, `hybris_types.h`) | Vendored. **Check its upstream licence before redistributing.** `android_shim.c`, `guard.cpp` and `dobby.h` (a stub) are original to this project |
| cJSON v1.7.19 | `drm/native/cJSON.*` | MIT, Dave Gamble and contributors |
| libhybris (prebuilt x86-64 build with a 10-line linker patch) | `drm/libhybris-core.so`, `drm/hybris-linker/q.so`, `drm/vendor/libhybris/` | Multi-licensed (Apache-2.0, LGPL-2.1, BSD, MIT, ISC, GPL-3.0 parts); licence texts, upstream commit, patch and rebuild script are in `drm/vendor/libhybris/` |
| Android and Apple Music native libraries | `drm/rootfs/system/lib64` at runtime | **Not distributed in this repository.** Apple's and Android's binaries are proprietary/third-party; install them yourself with `scripts/install-android-libs.sh` |
| `github.com/zhaarey/go-mp4tag` | Go dependency of `sdk/export` | MP4 tagging library |
| Other Go modules | `sdk/go.mod`, `server/go.mod` | See each module's licence |

"Apple", "Apple Music" and "FairPlay" are trademarks of Apple Inc. "Widevine" belongs to Google LLC. This project is not affiliated with either company.
