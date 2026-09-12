import Foundation
import Testing
@testable import GoreeCloudNetworkCore

@Test func decodesDevelopmentStatus() throws {
    let data = #"{"product":"GoreeCloud Network","version":"0.1.0-dev.4","lifecycle":"Development","surfaces":{"server":"development_identity_aware_control_plane","webDashboard":"development_dashboard","android":"development_client","googleTv":"development_client","ios":"development_client"}}"#.data(using: .utf8)!
    let status = try JSONDecoder().decode(NetworkStatus.self, from: data)
    #expect(status.product == "GoreeCloud Network")
    #expect(status.lifecycle == "Development")
    #expect(status.surfaces.server == "development_identity_aware_control_plane")
    #expect(status.surfaces.ios == "development_client")
}

@Test func decodesTruthfulOverview() throws {
    let data = #"{"deviceCount":0,"resourceCount":0,"policyCount":0,"policyMode":"deny_by_default","persistence":"development_file_store","schemaVersion":1,"revision":4,"migrationCount":1,"lastPersistedAt":"2026-09-11T23:00:00Z","integrity":"sha256","authentication":"identity_boundary_runtime_unconfigured"}"#.data(using: .utf8)!
    let overview = try JSONDecoder().decode(NetworkOverview.self, from: data)
    #expect(overview.deviceCount == 0)
    #expect(overview.policyMode == "deny_by_default")
    #expect(overview.persistence == "development_file_store")
    #expect(overview.schemaVersion == 1)
    #expect(overview.revision == 4)
    #expect(overview.integrity == "sha256")
    #expect(overview.authentication == "identity_boundary_runtime_unconfigured")
}
