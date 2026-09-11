plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
}

android {
    namespace = "com.goreecloud.network.tv"
    compileSdk = 37

    defaultConfig {
        applicationId = "com.goreecloud.network.tv"
        minSdk = 26
        targetSdk = 37
        versionCode = 1
        versionName = "0.1.0-dev.1"
    }

    buildTypes {
        getByName("debug") {
            manifestPlaceholders["usesCleartextTraffic"] = "true"
        }
        getByName("release") {
            manifestPlaceholders["usesCleartextTraffic"] = "false"
        }
    }
}

kotlin {
    jvmToolchain(17)
}
