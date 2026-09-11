#if canImport(SwiftUI)
import SwiftUI
import GoreeCloudNetworkCore

struct ContentView: View {
    @State private var statusText = "No live server state loaded. No tunnel is active."

    var body: some View {
        NavigationStack {
            List {
                Section("Development control plane") {
                    Text(statusText)
                    Button("Refresh Network state") { Task { await loadStatus() } }
                }
                Section("Current boundary") {
                    Text("The iOS client can inspect persisted Development control-plane revision metadata and deny-by-default decision state. Enrollment, authenticated administration, VPN tunneling, routing, relay, posture enforcement, and obfuscation remain unavailable.")
                }
            }
            .navigationTitle("GoreeCloud Network")
        }
    }

    @MainActor
    private func loadStatus() async {
        guard let url = URL(string: "http://127.0.0.1:8080/") else { return }
        do {
            let snapshot = try await NetworkAPI(baseURL: url).snapshot()
            statusText = "\(snapshot.status.product) \(snapshot.status.version)\n" +
                "Lifecycle: \(snapshot.status.lifecycle)\n" +
                "Server: \(snapshot.status.surfaces.server)\n" +
                "Devices: \(snapshot.overview.deviceCount)\n" +
                "Resources: \(snapshot.overview.resourceCount)\n" +
                "Policies: \(snapshot.overview.policyCount)\n" +
                "Policy mode: \(snapshot.overview.policyMode)\n" +
                "Persistence: \(snapshot.overview.persistence)\n" +
                "Schema: \(snapshot.overview.schemaVersion)\n" +
                "Revision: \(snapshot.overview.revision)\n" +
                "Migrations: \(snapshot.overview.migrationCount)\n" +
                "Integrity: \(snapshot.overview.integrity)\n" +
                "Authentication: \(snapshot.overview.authentication)"
        } catch {
            statusText = "Server unavailable: \(error.localizedDescription)\nNo connection state is being claimed."
        }
    }
}
#endif
