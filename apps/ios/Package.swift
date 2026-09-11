// swift-tools-version: 6.2
import PackageDescription

let package = Package(
    name: "GoreeCloudNetworkIOS",
    platforms: [.iOS(.v17)],
    products: [
        .library(name: "GoreeCloudNetworkCore", targets: ["GoreeCloudNetworkCore"]),
    ],
    targets: [
        .target(name: "GoreeCloudNetworkCore"),
        .testTarget(name: "GoreeCloudNetworkCoreTests", dependencies: ["GoreeCloudNetworkCore"]),
    ]
)
