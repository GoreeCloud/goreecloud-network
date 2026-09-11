import Foundation

public struct NetworkSurfaceStatus: Codable, Equatable, Sendable {
    public let server: String
    public let webDashboard: String
    public let android: String
    public let googleTv: String
    public let ios: String
}

public struct NetworkStatus: Codable, Equatable, Sendable {
    public let product: String
    public let version: String
    public let lifecycle: String
    public let surfaces: NetworkSurfaceStatus
}
