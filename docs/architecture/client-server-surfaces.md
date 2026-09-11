# Client and Server Surface Architecture

Server, Web Dashboard, Android, Google TV, and iOS are product surfaces of one GoreeCloud Network system. None may invent an independent capability state, authorization model, or lifecycle vocabulary.

The Server is the authoritative Network control-plane/API surface. The Web Dashboard renders server evidence rather than inferring protection. Android will own Android-specific VPN lifecycle and secure key storage. Google TV is a remote-first living-room surface. iOS will use Apple-native networking and secure-storage boundaries, with a future Network Extension tunnel provider.

All clients consume versioned API contracts. Breaking changes require explicit API versioning or compatibility migration rather than silent schema drift.
