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
        root.addView(TextView(this).apply { text = "Android client · Development bootstrap"; textSize = 16f })
        status = TextView(this).apply {
            text = "No live server state loaded. No tunnel is active."
            textSize = 18f
            setPadding(0, 48, 0, 32)
        }
        root.addView(status)
        root.addView(Button(this).apply {
            text = "Check local development server"
            setOnClickListener { loadStatus() }
        })
        root.addView(TextView(this).apply {
            text = "This bootstrap does not yet implement enrollment, VPN tunneling, routing, relay, policy enforcement, or obfuscation."
            setPadding(0, 48, 0, 0)
        })
        setContentView(root)
    }

    private fun loadStatus() {
        status.text = "Loading actual server state…"
        thread {
            val result = runCatching { NetworkApi("http://10.0.2.2:8080").status() }
            runOnUiThread {
                status.text = result.fold(
                    onSuccess = { "${it.product} ${it.version}\nLifecycle: ${it.lifecycle}\nServer: ${it.serverState}" },
                    onFailure = { "Server unavailable: ${it.message ?: "unknown error"}\nNo connection state is being claimed." },
                )
            }
        }
    }
}
