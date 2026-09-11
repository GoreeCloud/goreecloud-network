package com.goreecloud.network

import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL

data class NetworkStatus(
    val product: String,
    val version: String,
    val lifecycle: String,
    val serverState: String,
)

data class NetworkOverview(
    val deviceCount: Int,
    val resourceCount: Int,
    val policyCount: Int,
    val policyMode: String,
    val persistence: String,
    val authentication: String,
)

data class NetworkSnapshot(
    val status: NetworkStatus,
    val overview: NetworkOverview,
)

class NetworkApi(private val baseUrl: String) {
    fun snapshot(): NetworkSnapshot = NetworkSnapshot(status(), overview())

    fun status(): NetworkStatus {
        val json = getJson("/api/v1/status")
        return NetworkStatus(
            product = json.getString("product"),
            version = json.getString("version"),
            lifecycle = json.getString("lifecycle"),
            serverState = json.getJSONObject("surfaces").getString("server"),
        )
    }

    fun overview(): NetworkOverview {
        val json = getJson("/api/v1/overview")
        return NetworkOverview(
            deviceCount = json.getInt("deviceCount"),
            resourceCount = json.getInt("resourceCount"),
            policyCount = json.getInt("policyCount"),
            policyMode = json.getString("policyMode"),
            persistence = json.getString("persistence"),
            authentication = json.getString("authentication"),
        )
    }

    private fun getJson(path: String): JSONObject {
        val connection = URL("${baseUrl.trimEnd('/')}$path").openConnection() as HttpURLConnection
        connection.requestMethod = "GET"
        connection.connectTimeout = 5_000
        connection.readTimeout = 5_000
        connection.setRequestProperty("Accept", "application/json")
        return try {
            if (connection.responseCode !in 200..299) error("HTTP ${connection.responseCode}")
            val body = connection.inputStream.bufferedReader().use { it.readText() }
            JSONObject(body)
        } finally {
            connection.disconnect()
        }
    }
}
