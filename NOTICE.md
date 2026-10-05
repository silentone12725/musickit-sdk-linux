# Third-party code and credits

This project is MIT licensed (see `LICENSE`). It includes or depends on the following; each retains its own licence.

| Component | Where | Notes |
|---|---|---|
| Reference host-native DRM wrapper (WorldObservationLog/zhaarey wrapper lineage, with local lease-recovery and library-mode changes) | `drm/native/` (`main.c`, `main.cpp`, `hybris_stubs.c`, `drm_lib.c`, `cmdline.*`, `import.h`, `hybris_types.h`) | Vendored. **Check its upstream licence before redistributing.** `android_shim.c`, `guard.cpp` and `dobby.h` (a stub) are original to this project |
| cJSON v1.7.19 | `drm/native/cJSON.*` | MIT, Dave Gamble and contributors |
| libhybris | linked at runtime / embedded into `libdrm_client.so` by the Makefile | Not included in this repository; see its own licence |
| Android and Apple Music native libraries | runtime, supplied by you | Not distributed here |
| `github.com/zhaarey/go-mp4tag` | Go dependency of `sdk/export` | MP4 tagging library |
| Other Go modules | `sdk/go.mod`, `server/go.mod` | See each module's licence |

"Apple", "Apple Music" and "FairPlay" are trademarks of Apple Inc. "Widevine" belongs to Google LLC. This project is not affiliated with either company.
