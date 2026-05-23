package ro.ssproject.medicalocrclient

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import ro.ssproject.medicalocrclient.ui.theme.SSMedicalOCRClientTheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        setContent {
            SSMedicalOCRClientTheme {
                MedicalOcrClientScreen()
            }
        }
    }
}

@Composable
fun MedicalOcrClientScreen() {
    Scaffold(
        modifier = Modifier.fillMaxSize()
    ) { innerPadding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding)
                .padding(20.dp),
            verticalArrangement = Arrangement.Top
        ) {
            Text(
                text = "SS Medical OCR Client",
                style = MaterialTheme.typography.headlineSmall
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = "Capture or select a medical document image and send it to the OCR backend through MQTT.",
                style = MaterialTheme.typography.bodyMedium
            )

            Spacer(modifier = Modifier.height(24.dp))

            Button(
                onClick = {
                    // TODO: open camera
                },
                modifier = Modifier.fillMaxWidth()
            ) {
                Text("Take photo")
            }

            Spacer(modifier = Modifier.height(12.dp))

            Button(
                onClick = {
                    // TODO: open gallery picker
                },
                modifier = Modifier.fillMaxWidth()
            ) {
                Text("Select from gallery")
            }

            Spacer(modifier = Modifier.height(24.dp))

            Text(
                text = "Upload queue",
                style = MaterialTheme.typography.titleMedium
            )

            Spacer(modifier = Modifier.height(12.dp))

            UploadStatusCard(
                fileName = "No images yet",
                status = "Waiting for capture"
            )
        }
    }
}

@Composable
fun UploadStatusCard(
    fileName: String,
    status: String
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            Text(
                text = fileName,
                style = MaterialTheme.typography.bodyLarge
            )

            Spacer(modifier = Modifier.height(4.dp))

            Text(
                text = status,
                style = MaterialTheme.typography.bodyMedium
            )
        }
    }
}