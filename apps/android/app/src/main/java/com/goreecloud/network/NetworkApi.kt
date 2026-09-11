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

class NetworkApi(private val baseUrl: String) {
    fun status(): NetworkStatus {
        val connection = URL("${baseUrl.trimEnd('/')}/api/v1/status").openConnection() as HttpURLConnection
        connection.requestMethod = "GET"
        connection.connectTimeout = 5_000
        connection.readTimeout = 5_000
        connection.setRequestProperty("Accept", "application/json")
        return try {
            if (connection.responseCode !in 200..299) error("HTTP ${connection.responseCode}")
            val body = connection.inputStream.bufferedReader().use { it.readText() }
            val json = JSONObject(body)
            NetworkStatus(
                product = json.getString("product"),
                version = json.getString("version"),
                lifecycle = json.getString("lifecycle"),
                serverState = json.getJSONObject("surfaces").getString("server"),
            )
        } finally {
            connection.disconnect()
        }
    }
}
