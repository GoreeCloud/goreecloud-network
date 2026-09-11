# Repository Notes

- Lifecycle: Development.
- Version: `0.1.0-dev.1`.
- Native repository origin: created under GoreeCloud; the connected `main` history began with a GoreeCloud initial commit rather than an upstream product fork.
- Current default branch before this bootstrap contained only a two-line README.
- Android and Google TV build files target API 36 and use Android Gradle Plugin 9.3.1 built-in Kotlin support with Gradle 9.5.0; platform CI still requires execution on an Android SDK runner.
- iOS source requires an Apple SDK with SwiftUI; portable core tests validate the shared Swift package, while a signed iOS application build remains a later platform acceptance requirement.
- A repository license is included as GPL-3.0 based on the current GoreeCloud open-source requirement and existing GoreeCloud native-repository precedent; licensing must remain subject to GoreeCloud legal/governance review.
- Platform System adapters are not considered implemented merely because this repository declares their planned boundaries.
