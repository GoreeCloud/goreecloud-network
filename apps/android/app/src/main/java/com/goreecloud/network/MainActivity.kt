package com.goreecloud.network

import android.app.Activity
import android.os.Bundle
import android.view.Gravity
import android.widget.Button
import android.widget.LinearLayout
import android.widget.TextView
import kotlin.concurrent.thread

class MainActivity : Activity() {
    private lateinit var status: TextView

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val root = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER_HORIZONTAL
            setPadding(48, 64, 48, 48)
        }
        root.addView(TextView(this).apply { text = "GoreeCloud Network"; textSize = 28f })
        root.addView(TextView(this).apply { text = "Android client · Development"; textSize = 16f })
        status = TextView(this).apply {
            text = "No live server state loaded. No tunnel is active."
            textSize = 18f
            setPadding(0, 48, 0, 32)
        }
        root.addView(status)
        root.addView(Button(this).apply {
            text = "Refresh control-plane state"
            setOnClickListener { loadStatus() }
        })
        root.addView(TextView(this).apply {
            text = "The client can inspect the native Development control-plane inventory and policy mode. Enrollment, VPN tunneling, routing, relay, durable policy administration, and obfuscation are not yet implemented."
            setPadding(0, 48, 0, 0)
        })
        setContentView(root)
    }

    private fun loadStatus() {
        status.text = "Loading actual server state…"
        thread {
            val result = runCatching { NetworkApi("http://10.0.2.2:8080").snapshot() }
            runOnUiThread {
                status.text = result.fold(
                    onSuccess = {
                        "${it.status.product} ${it.status.version}\n" +
                            "Lifecycle: ${it.status.lifecycle}\n" +
                            "Server: ${it.status.serverState}\n" +
                            "Devices: ${it.overview.deviceCount}\n" +
                            "Resources: ${it.overview.resourceCount}\n" +
                            "Policies: ${it.overview.policyCount}\n" +
                            "Policy mode: ${it.overview.policyMode}\n" +
                            "Persistence: ${it.overview.persistence}\n" +
                            "Authentication: ${it.overview.authentication}"
                    },
                    onFailure = { "Server unavailable: ${it.message ?: "unknown error"}\nNo connection state is being claimed." },
                )
            }
        }
    }
}
