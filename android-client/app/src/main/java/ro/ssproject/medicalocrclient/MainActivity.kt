package ro.ssproject.medicalocrclient

import android.Manifest
import android.content.Context
import android.net.Uri
import android.os.Bundle
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.compose.setContent
import androidx.activity.result.contract.ActivityResultContracts
import androidx.activity.result.PickVisualMediaRequest
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
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
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.core.content.FileProvider
import ro.ssproject.medicalocrclient.ui.theme.SSMedicalOCRClientTheme
import java.io.File
import java.io.FileOutputStream
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import androidx.compose.runtime.rememberCoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

data class UploadItem(
    val fileName: String,
    val filePath: String,
    val status: String
)

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
    val context = LocalContext.current
    val uploadItems = remember { mutableStateListOf<UploadItem>() }

    val coroutineScope = rememberCoroutineScope()

    var currentCameraUri by remember { mutableStateOf<Uri?>(null) }
    var currentCameraFile by remember { mutableStateOf<File?>(null) }

    val takePictureLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.TakePicture()
    ) { success ->
        val imageFile = currentCameraFile

        if (success && imageFile != null && imageFile.exists()) {
            uploadItems.add(
                0,
                UploadItem(
                    fileName = imageFile.name,
                    filePath = imageFile.absolutePath,
                    status = "PENDING"
                )
            )

            Toast.makeText(context, "Photo saved as pending upload", Toast.LENGTH_SHORT).show()
        } else {
            Toast.makeText(context, "Photo capture cancelled or failed", Toast.LENGTH_SHORT).show()
        }
    }

    val cameraPermissionLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestPermission()
    ) { granted ->
        if (granted) {
            val imageFile = createImageFile(context)
            val imageUri = FileProvider.getUriForFile(
                context,
                "${context.packageName}.fileprovider",
                imageFile
            )

            currentCameraFile = imageFile
            currentCameraUri = imageUri
            takePictureLauncher.launch(imageUri)
        } else {
            Toast.makeText(context, "Camera permission is required", Toast.LENGTH_SHORT).show()
        }
    }

    val galleryLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.PickVisualMedia()
    ) { uri ->
        if (uri != null) {
            val copiedFile = copyGalleryImageToLocalStorage(context, uri)

            if (copiedFile != null) {
                uploadItems.add(
                    0,
                    UploadItem(
                        fileName = copiedFile.name,
                        filePath = copiedFile.absolutePath,
                        status = "PENDING"
                    )
                )

                Toast.makeText(context, "Image selected as pending upload", Toast.LENGTH_SHORT).show()
            } else {
                Toast.makeText(context, "Failed to copy selected image", Toast.LENGTH_SHORT).show()
            }
        }
    }

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
                    cameraPermissionLauncher.launch(Manifest.permission.CAMERA)
                },
                modifier = Modifier.fillMaxWidth()
            ) {
                Text("Take photo")
            }

            Spacer(modifier = Modifier.height(12.dp))

            Button(
                onClick = {
                    galleryLauncher.launch(
                        PickVisualMediaRequest(ActivityResultContracts.PickVisualMedia.ImageOnly)
                    )
                },
                modifier = Modifier.fillMaxWidth()
            ) {
                Text("Select from gallery")
            }


            Spacer(modifier = Modifier.height(12.dp))

            Button(
                onClick = {
                    coroutineScope.launch {
                        val pendingIndexes = uploadItems
                            .mapIndexedNotNull { index, item ->
                                if (item.status == "PENDING" || item.status == "FAILED") index else null
                            }

                        if (pendingIndexes.isEmpty()) {
                            Toast.makeText(context, "No pending images to upload", Toast.LENGTH_SHORT).show()
                            return@launch
                        }

                        pendingIndexes.forEach { index ->
                            val item = uploadItems[index]

                            uploadItems[index] = item.copy(status = "UPLOADING")

                            try {
                                withContext(Dispatchers.IO) {
                                    MqttUploader.uploadImage(context, File(item.filePath))
                                }

                                uploadItems[index] = item.copy(status = "SENT")
                                Toast.makeText(context, "Uploaded: ${item.fileName}", Toast.LENGTH_SHORT).show()
                            } catch (e: Exception) {
                                e.printStackTrace()
                                uploadItems[index] = item.copy(status = "FAILED")
                                Toast.makeText(
                                    context,
                                    "Upload failed: ${e.message}",
                                    Toast.LENGTH_LONG
                                ).show()
                            }
                        }
                    }
                },
                modifier = Modifier.fillMaxWidth()
            ) {
                Text("Upload pending images")
            }

            Spacer(modifier = Modifier.height(24.dp))

            Text(
                text = "Upload queue",
                style = MaterialTheme.typography.titleMedium
            )

            Spacer(modifier = Modifier.height(12.dp))

            if (uploadItems.isEmpty()) {
                UploadStatusCard(
                    fileName = "No images yet",
                    filePath = "-",
                    status = "Waiting for capture"
                )
            } else {
                uploadItems.forEach { item ->
                    UploadStatusCard(
                        fileName = item.fileName,
                        filePath = item.filePath,
                        status = item.status
                    )

                    Spacer(modifier = Modifier.height(10.dp))
                }
            }
        }
    }
}

@Composable
fun UploadStatusCard(
    fileName: String,
    filePath: String,
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

            Row {
                Text(
                    text = "Status: ",
                    style = MaterialTheme.typography.bodyMedium
                )

                Text(
                    text = status,
                    style = MaterialTheme.typography.bodyMedium
                )
            }

            Spacer(modifier = Modifier.height(4.dp))

            Text(
                text = filePath,
                style = MaterialTheme.typography.bodySmall
            )
        }
    }
}

fun createImageFile(context: Context): File {
    val timestamp = SimpleDateFormat("yyyyMMdd_HHmmss", Locale.US).format(Date())
    val directory = File(context.cacheDir, "captured_images")

    if (!directory.exists()) {
        directory.mkdirs()
    }

    return File(directory, "medical_doc_$timestamp.jpg")
}

fun copyGalleryImageToLocalStorage(context: Context, uri: Uri): File? {
    return try {
        val timestamp = SimpleDateFormat("yyyyMMdd_HHmmss", Locale.US).format(Date())
        val directory = File(context.filesDir, "stored_images")

        if (!directory.exists()) {
            directory.mkdirs()
        }

        val destinationFile = File(directory, "gallery_doc_$timestamp.jpg")

        context.contentResolver.openInputStream(uri)?.use { input ->
            FileOutputStream(destinationFile).use { output ->
                input.copyTo(output)
            }
        }

        destinationFile
    } catch (e: Exception) {
        e.printStackTrace()
        null
    }
}