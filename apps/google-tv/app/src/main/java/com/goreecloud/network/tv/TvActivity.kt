package com.goreecloud.network.tv

import android.app.Activity
import android.os.Bundle
import android.view.Gravity
import android.widget.Button
import android.widget.LinearLayout
import android.widget.TextView
import kotlin.concurrent.thread

class TvActivity : Activity() {
    private lateinit var status: TextView

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val root = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER
            setPadding(72, 72, 72, 72)
        }
        root.addView(TextView(this).apply { text = "GoreeCloud Network"; textSize = 36f })
        root.addView(TextView(this).apply { text = "Google TV client · Development"; textSize = 20f })
        status = TextView(this).apply {
            text = "No live server state loaded. No tunnel is active."
            textSize = 24f
            setPadding(0, 56, 0, 32)
        }
        root.addView(status)
        root.addView(Button(this).apply {
            text = "Refresh Network state"
            isFocusable = true
            setOnClickListener { loadStatus() }
        })
        root.addView(TextView(this).apply {
            text = "This TV client currently presents verified Development control-plane state only. Remote-first connection controls will be added after enrollment and tunnel runtime implementation exists."
            textSize = 18f
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
                        "Version ${it.status.version}\n" +
                            "Lifecycle: ${it.status.lifecycle}\n" +
                            "Server: ${it.status.serverState}\n" +
                            "Devices: ${it.overview.deviceCount}\n" +
                            "Resources: ${it.overview.resourceCount}\n" +
                            "Policies: ${it.overview.policyCount}\n" +
                            "Policy mode: ${it.overview.policyMode}"
                    },
                    onFailure = { "Server unavailable: ${it.message ?: "unknown error"}\nNo connection state is being claimed." },
                )
            }
        }
    }
}
