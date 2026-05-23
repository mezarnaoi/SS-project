package ro.ssproject.medicalocrclient

import android.content.Context
import org.eclipse.paho.client.mqttv3.MqttClient
import org.eclipse.paho.client.mqttv3.MqttConnectOptions
import org.eclipse.paho.client.mqttv3.MqttMessage
import org.eclipse.paho.client.mqttv3.persist.MemoryPersistence
import java.io.File
import java.security.KeyStore
import java.security.SecureRandom
import java.security.cert.CertificateFactory
import javax.net.ssl.KeyManagerFactory
import javax.net.ssl.SSLContext
import javax.net.ssl.SSLSocketFactory
import javax.net.ssl.TrustManagerFactory

object MqttUploader {
    private const val MQTT_SERVER_URI = "ssl://10.0.2.2:8883"
    private const val DEVICE_ID = "android-emulator-1"
    private const val P12_PASSWORD = "changeit"

    private const val REGISTER_TOPIC = "register/$DEVICE_ID"
    private const val PHOTO_TOPIC = "ssproject/images/$DEVICE_ID"

    fun uploadImage(context: Context, imageFile: File) {
        if (!imageFile.exists()) {
            throw IllegalArgumentException("Image file does not exist: ${imageFile.absolutePath}")
        }

        val clientId = "android-client-${System.currentTimeMillis()}"
        val client = MqttClient(MQTT_SERVER_URI, clientId, MemoryPersistence())

        val options = MqttConnectOptions().apply {
            isCleanSession = true
            socketFactory = createSocketFactory(context)
            connectionTimeout = 10
            keepAliveInterval = 30
        }

        try {
            client.connect(options)

            val registerPayload = """
                {
                  "device_id": "$DEVICE_ID",
                  "type": "android",
                  "status": "online"
                }
            """.trimIndent().toByteArray(Charsets.UTF_8)

            client.publish(
                REGISTER_TOPIC,
                MqttMessage(registerPayload).apply {
                    qos = 1
                    isRetained = false
                }
            )

            val imageBytes = imageFile.readBytes()

            client.publish(
                PHOTO_TOPIC,
                MqttMessage(imageBytes).apply {
                    qos = 1
                    isRetained = false
                }
            )
        } finally {
            if (client.isConnected) {
                client.disconnect()
            }
            client.close()
        }
    }

    fun createSocketFactory(context: Context): SSLSocketFactory {

        val certificateFactory = CertificateFactory.getInstance("X.509")

        val caCertificate = context.resources.openRawResource(R.raw.ca).use { input ->
            certificateFactory.generateCertificate(input)
        }

        val trustStore = KeyStore.getInstance(KeyStore.getDefaultType()).apply {
            load(null, null)
            setCertificateEntry("ca", caCertificate)
        }

        val trustManagerFactory = TrustManagerFactory.getInstance(
            TrustManagerFactory.getDefaultAlgorithm()
        ).apply {
            init(trustStore)
        }

        val keyStore = KeyStore.getInstance("PKCS12").apply {
            context.resources.openRawResource(R.raw.android_client).use { input ->
                load(input, P12_PASSWORD.toCharArray())
            }
        }

        val keyManagerFactory = KeyManagerFactory.getInstance(
            KeyManagerFactory.getDefaultAlgorithm()
        ).apply {
            init(keyStore, P12_PASSWORD.toCharArray())
        }

        val sslContext = SSLContext.getInstance("TLS").apply {
            init(
                keyManagerFactory.keyManagers,
                trustManagerFactory.trustManagers,
                SecureRandom()
            )
        }

        return sslContext.socketFactory
    }
}