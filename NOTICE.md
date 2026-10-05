# Third-party code and credits

This project is MIT licensed (see `LICENSE`). It includes or depends on the following; each retains its own licence.

| Component | Where | Notes |
|---|---|---|
| Reference host-native DRM wrapper (WorldObservationLog/zhaarey wrapper lineage, with local lease-recovery and library-mode changes) | `drm/native/` (`main.c`, `main.cpp`, `hybris_stubs.c`, `drm_lib.c`, `cmdline.*`, `import.h`, `hybris_types.h`) | Vendored. Upstream (WorldObservationLog/wrapper) is MIT-licensed; its licence text is in `drm/native/LICENSE.wrapper` and must accompany redistributions. `cJSON.*` is MIT, text in `drm/native/LICENSE.cJSON`. `android_shim.c`, `guard.cpp` and `dobby.h` (a stub) are original to this project |
| cJSON v1.7.19 | `drm/native/cJSON.*` | MIT, Dave Gamble and contributors |
| libhybris (prebuilt x86-64 build with a 10-line linker patch) | `drm/libhybris-core.so`, `drm/hybris-linker/q.so`, `drm/vendor/libhybris/` | Multi-licensed (Apache-2.0, LGPL-2.1, BSD, MIT, ISC, GPL-3.0 parts); licence texts, upstream commit, patch and rebuild script are in `drm/vendor/libhybris/` |
| Android and Apple Music native libraries (bionic libc/libm/libdl and a few basic system libraries, Apple Music store services, FairPlay and media libraries, ICU; 23 files, listed in `drm/android-libs.txt`) | `drm/rootfs/system/lib64` | **Proprietary / third-party binaries included as prebuilt files.** They remain the property of Apple Inc., Google/AOSP contributors and their respective licensors and are not covered by this project's MIT licence. Confirm you have the right to redistribute them before publishing this repository; `scripts/install-android-libs.sh` can replace them with a set you supply |
| `github.com/zhaarey/go-mp4tag` | Go dependency of `sdk/export` | MP4 tagging library |
| Other Go modules | `sdk/go.mod`, `server/go.mod` | See each module's licence |

"Apple", "Apple Music" and "FairPlay" are trademarks of Apple Inc. "Widevine" belongs to Google LLC. This project is not affiliated with either company.
