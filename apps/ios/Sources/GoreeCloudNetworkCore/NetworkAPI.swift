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

    public func status() async throws -> NetworkStatus {
        let url = baseURL.appending(path: "api/v1/status")
        var request = URLRequest(url: url)
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.timeoutInterval = 5
        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse, (200..<300).contains(http.statusCode) else {
            throw URLError(.badServerResponse)
        }
        return try JSONDecoder().decode(NetworkStatus.self, from: data)
    }
}
