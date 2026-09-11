plugins {
    id("com.android.application")
}

android {
    namespace = "com.goreecloud.network"
    compileSdk = 36

    defaultConfig {
        applicationId = "com.goreecloud.network"
        minSdk = 26
        targetSdk = 36
        versionCode = 1
        versionName = "0.1.0-dev.1"
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
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
