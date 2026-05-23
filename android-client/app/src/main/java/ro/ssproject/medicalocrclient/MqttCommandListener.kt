package ro.ssproject.medicalocrclient

import android.content.Context
import org.eclipse.paho.client.mqttv3.IMqttMessageListener
import org.eclipse.paho.client.mqttv3.MqttClient
import org.eclipse.paho.client.mqttv3.MqttConnectOptions
import org.eclipse.paho.client.mqttv3.persist.MemoryPersistence
import org.json.JSONObject
import android.util.Log

class MqttCommandListener(
    private val context: Context,
    private val onCaptureCommand: () -> Unit,
    private val onStartLiveCommand: () -> Unit,
    private val onStopLiveCommand: () -> Unit
) {
    private var client: MqttClient? = null

    companion object {
        private const val MQTT_SERVER_URI = "ssl://10.0.2.2:8883"
        private const val DEVICE_ID = "android-emulator-1"
        private const val COMMAND_TOPIC = "ssproject/commands"
    }

    fun connectAndSubscribe() {
        val clientId = "android-command-listener-${System.currentTimeMillis()}"
        val mqttClient = MqttClient(MQTT_SERVER_URI, clientId, MemoryPersistence())

        val options = MqttConnectOptions().apply {
            isCleanSession = true
            socketFactory = MqttUploader.createSocketFactory(context)
            connectionTimeout = 10
            keepAliveInterval = 30
            isAutomaticReconnect = true
        }

        mqttClient.connect(options)

        mqttClient.subscribe(COMMAND_TOPIC, 1, IMqttMessageListener { _, message ->
            val payload = message.payload.toString(Charsets.UTF_8)
            Log.d("MqttCommandListener", "Received MQTT command payload: [$payload]")
            handleCommand(payload)
        })

        client = mqttClient
    }

    fun disconnect() {
        client?.let {
            if (it.isConnected) {
                it.disconnect()
            }
            it.close()
        }
        client = null
    }

    private fun handleCommand(payload: String) {
        try {
            val trimmedPayload = payload.trim()
            Log.d("MqttCommandListener", "Handling command payload: [$trimmedPayload]")

            val command = if (trimmedPayload.startsWith("{")) {
                val json = JSONObject(trimmedPayload)

                val target = json.optString("target", "all")
                if (target != DEVICE_ID && target != "all") {
                    Log.d("MqttCommandListener", "Ignoring command for target: $target")
                    return
                }

                json.optString("command")
            } else {
                trimmedPayload
            }.lowercase()

            Log.d("MqttCommandListener", "Parsed command: [$command]")

            when (command) {
                "capture" -> {
                    Log.d("MqttCommandListener", "Executing capture callback")
                    onCaptureCommand()
                }

                "start_live" -> {
                    Log.d("MqttCommandListener", "Executing start_live callback")
                    onStartLiveCommand()
                }

                "stop_live" -> {
                    Log.d("MqttCommandListener", "Executing stop_live callback")
                    onStopLiveCommand()
                }

                else -> {
                    Log.w("MqttCommandListener", "Unknown MQTT command received: [$command]")
                }
            }
        } catch (e: Exception) {
            Log.e("MqttCommandListener", "Failed to handle MQTT command", e)
        }
    }
}