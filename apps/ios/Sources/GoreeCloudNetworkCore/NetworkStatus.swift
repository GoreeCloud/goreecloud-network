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

public struct NetworkOverview: Codable, Equatable, Sendable {
    public let deviceCount: Int
    public let resourceCount: Int
    public let policyCount: Int
    public let policyMode: String
    public let persistence: String
    public let schemaVersion: Int
    public let revision: UInt64
    public let migrationCount: Int
    public let lastPersistedAt: String?
    public let integrity: String
    public let authentication: String
}

public struct NetworkSnapshot: Equatable, Sendable {
    public let status: NetworkStatus
    public let overview: NetworkOverview
}
