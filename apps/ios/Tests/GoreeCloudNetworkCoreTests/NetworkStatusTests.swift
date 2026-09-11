import Foundation
import Testing
@testable import GoreeCloudNetworkCore

@Test func decodesDevelopmentStatus() throws {
    let data = #"{"product":"GoreeCloud Network","version":"0.1.0-dev.1","lifecycle":"Development","surfaces":{"server":"bootstrap","webDashboard":"bootstrap","android":"source_bootstrap","googleTv":"source_bootstrap","ios":"source_bootstrap"}}"#.data(using: .utf8)!
    let status = try JSONDecoder().decode(NetworkStatus.self, from: data)
    #expect(status.product == "GoreeCloud Network")
    #expect(status.lifecycle == "Development")
    #expect(status.surfaces.ios == "source_bootstrap")
}
