#if canImport(SwiftUI)
import SwiftUI
import GoreeCloudNetworkCore

struct ContentView: View {
    @State private var statusText = "No live server state loaded. No tunnel is active."

    var body: some View {
        NavigationStack {
            List {
                Section("Development status") {
                    Text(statusText)
                    Button("Check local development server") { Task { await loadStatus() } }
                }
                Section("Not yet implemented") {
                    Text("Enrollment, VPN tunneling, routing, relay, access-policy enforcement, posture, and obfuscation remain unavailable until verified source and runtime implementation exists.")
                }
            }
            .navigationTitle("GoreeCloud Network")
        }
    }

    @MainActor
    private func loadStatus() async {
        guard let url = URL(string: "http://127.0.0.1:8080/") else { return }
        do {
            let status = try await NetworkAPI(baseURL: url).status()
            statusText = "\(status.product) \(status.version)\nLifecycle: \(status.lifecycle)\nServer: \(status.surfaces.server)"
        } catch {
            statusText = "Server unavailable: \(error.localizedDescription)\nNo connection state is being claimed."
        }
    }
}
#endif
