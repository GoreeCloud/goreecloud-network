import Foundation
#if canImport(FoundationNetworking)
import FoundationNetworking
#endif

public actor NetworkAPI {
    private let baseURL: URL
    private let session: URLSession

    public init(baseURL: URL, session: URLSession = .shared) {
        self.baseURL = baseURL
        self.session = session
    }

    public func snapshot() async throws -> NetworkSnapshot {
        async let status = status()
        async let overview = overview()
        return try await NetworkSnapshot(status: status, overview: overview)
    }

    public func status() async throws -> NetworkStatus {
        try await get(path: "api/v1/status", as: NetworkStatus.self)
    }

    public func overview() async throws -> NetworkOverview {
        try await get(path: "api/v1/overview", as: NetworkOverview.self)
    }

    private func get<T: Decodable & Sendable>(path: String, as type: T.Type) async throws -> T {
        let url = baseURL.appendingPathComponent(path)
        var request = URLRequest(url: url)
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.timeoutInterval = 5
        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse, (200..<300).contains(http.statusCode) else {
            throw URLError(.badServerResponse)
        }
        return try JSONDecoder().decode(type, from: data)
    }
}
