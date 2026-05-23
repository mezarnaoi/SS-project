# Android Mobile Client

This folder contains the Android mobile client for the **SS secure medical OCR platform**.

The app is intentionally simple and is meant to satisfy the mobile/client-side requirement of the project. It acts as a real Android client that captures or selects a medical document image and sends it to the backend through **MQTT over mutual TLS**.

The Android client integrates with the existing backend by publishing to the same MQTT topic structure already used by the Python test client.

---

## Features

The Android client currently supports:

- Android/Kotlin application built with Android Studio
- Jetpack Compose UI
- image capture using the Android emulator camera
- optional image selection from gallery
- local pending upload queue display
- MQTT image upload using Eclipse Paho
- MQTT over mutual TLS
- device registration message
- image upload to the topic expected by the Go backend
- upload status display:
  - `PENDING`
  - `UPLOADING`
  - `SENT`
  - `FAILED`

---

## Integration with the backend

The backend already receives medical document images through MQTT.

The Android client publishes to the same topic structure as the existing Python test client:

```text
register/{deviceId}
ssproject/images/{deviceId}
```

For the current emulator setup, the Android device ID is:

```text
android-emulator-1
```

Therefore, the Android app publishes:

```text
register/android-emulator-1
ssproject/images/android-emulator-1
```

The expected flow is:

```text
Android Emulator
      |
      | MQTT over mTLS
      v
Mosquitto Broker
      |
      v
Go Backend
      |
      v
OCR Sandbox
      |
      v
Structured medical data extraction / storage
```

---

## Emulator networking

When the app runs inside the Android Studio emulator, `localhost` does not refer to the host laptop.

Inside the emulator:

```text
localhost / 127.0.0.1 = the emulator itself
```

To reach services running on the host laptop, Android uses:

```text
10.0.2.2
```

Therefore, the Android app connects to the MQTT broker using:

```text
ssl://10.0.2.2:8883
```

Those would point to the emulator, not to the broker running on the host machine.

---

## TLS and mTLS overview

The project uses **mutual TLS** for the MQTT connection.

Normal TLS only authenticates the server:

```text
Client verifies server
```

Mutual TLS authenticates both sides:

```text
Client verifies server
Server verifies client
```

In this project:

- the broker presents a server certificate
- the Android app verifies the broker using the project CA certificate
- the Android app presents a client certificate
- the broker verifies the Android client certificate using the same project CA

---

## Certificate files

The project certificate material is generated under the root-level `secrets/` folder.

Typical files are:

```text
secrets/ca.crt
secrets/ca.key
secrets/ca.srl
secrets/server.crt
secrets/server.key
secrets/web.crt
secrets/web.key
```

### `ca.crt`

This is the public certificate of the project Certificate Authority.

It is used by the Android app to verify the MQTT broker certificate.

This file is public, but it is generated locally and environment-specific.

### `ca.key`

This is the private key of the Certificate Authority.

This file is sensitive and must not be copied into the Android app.

It must not be committed.

### `server.crt`

This is the certificate presented by the Mosquitto broker.

The Android app verifies this certificate during the TLS handshake.

### `server.key`

This is the private key corresponding to `server.crt`.

It is used by the broker and must not be copied into the Android app.

### `web.crt`

This is the client certificate used by the existing Python test client.

For the Android client, this certificate is reused as the client certificate.

### `web.key`

This is the private key corresponding to `web.crt`.

Android needs this private key for mTLS client authentication, but it is not copied as a raw `.key` file. Instead, it is packaged into a PKCS#12 keystore.

---

## Required local Android TLS files

The Android app expects the following files to exist locally:

```text
android-client/app/src/main/res/raw/ca.crt
android-client/app/src/main/res/raw/android_client.p12
```

These files are intentionally not committed to Git.

The file `android_client.p12` contains the Android client's certificate and private key, so it must remain local.

---

## Why `android_client.p12` is needed

Python can load certificate files like this:

```text
ca.crt
web.crt
web.key
```

Android/Java typically uses a keystore format instead.

For this reason, the Android client uses a PKCS#12 file:

```text
android_client.p12
```

This file contains:

```text
web.crt
web.key
ca.crt certificate chain
```

Conceptually:

```text
secrets/web.crt + secrets/web.key + secrets/ca.crt
        |
        v
android_client.p12
        |
        v
Android app mTLS client authentication
```

---

## Broker certificate SAN requirement

Because the Android emulator connects to the host machine through:

```text
10.0.2.2
```

the broker certificate must include `10.0.2.2` as a Subject Alternative Name.

The certificate generation script should include entries similar to:

```text
DNS.1 = broker
DNS.2 = localhost
IP.1 = 127.0.0.1
IP.2 = 10.0.2.2
```

If `10.0.2.2` is missing from the broker certificate, Android may reject the broker certificate even if it was signed by the correct CA.

After editing the certificate generation script, certificates must be regenerated. Editing the script alone does not modify already generated certificates.

---

## Generating certificates

From the project root, run:

```bash
./scripts/gen-certs.sh
```

After generating the certificates, verify that the broker certificate contains the emulator IP:

```bash
openssl x509 -in secrets/server.crt -noout -text | grep -A 3 "Subject Alternative Name"
```

Expected entries should include:

```text
DNS:broker
DNS:localhost
IP Address:127.0.0.1
IP Address:10.0.2.2
```

On Windows Git Bash, the command above should work.

If using PowerShell and `grep` is unavailable, use:

```powershell
openssl x509 -in secrets/server.crt -noout -text
```

Then manually search the output for:

```text
Subject Alternative Name
```

---

## Creating the Android PKCS#12 keystore

From the project root, create the Android keystore:

```bash
openssl pkcs12 -export \
  -in secrets/web.crt \
  -inkey secrets/web.key \
  -certfile secrets/ca.crt \
  -out android_client.p12 \
  -name android-client \
  -passout pass:changeit
```

This creates:

```text
android_client.p12
```

The password used here is:

```text
changeit
```

The Android code must use the same password when loading the keystore.

In `MqttUploader.kt`, the password is configured as:

```kotlin
private const val P12_PASSWORD = "changeit"
```

If another password is used when generating the `.p12`, this value must be updated.

---

## Copying TLS files into the Android app

After generating the certificates and the Android PKCS#12 keystore, copy the files into the Android project.

From the project root, using Git Bash:

```bash
cp android_client.p12 android-client/app/src/main/res/raw/android_client.p12
cp secrets/ca.crt android-client/app/src/main/res/raw/ca.crt
```

On Windows PowerShell:

```powershell
Copy-Item .\android_client.p12 .\android-client\app\src\main\res\raw\android_client.p12 -Force
Copy-Item .\secrets\ca.crt .\android-client\app\src\main\res\raw\ca.crt -Force
```

The final local structure should be:

```text
android-client/app/src/main/res/raw/ca.crt
android-client/app/src/main/res/raw/android_client.p12
```

---

## Starting the backend stack

From the project root:

```bash
docker compose up -d
```

Check that the broker is running:

```bash
docker ps
```

The MQTT broker must expose port `8883`.

If the certificates were regenerated, restart the broker so that it reloads the new certificate files:

```bash
docker compose restart broker
```

Alternatively, recreate the full stack:

```bash
docker compose down
docker compose up -d
```

---

## Running the Android app

Open the following folder in Android Studio:

```text
android-client
```

Run the app on an Android emulator.

Expected flow:

1. Start the backend stack with Docker Compose.
2. Start the Android app in the emulator.
3. Press `Take photo` or `Select from gallery`.
4. The image appears in the upload queue as `PENDING`.
5. Press `Upload pending images`.
6. The app connects to the broker using MQTT over mTLS.
7. The app publishes a registration message to:

```text
register/android-emulator-1
```

8. The app publishes the image bytes to:

```text
ssproject/images/android-emulator-1
```

9. If the upload succeeds, the status changes to:

```text
SENT
```

10. If the broker is unavailable or TLS validation fails, the status changes to:

```text
FAILED
```

---

## Building from command line

From the Android client folder:

```bash
cd android-client
./gradlew assembleDebug
```

On Windows PowerShell:

```powershell
cd android-client
.\gradlew assembleDebug
```

If Gradle cannot delete build directories on Windows, stop Gradle and close Android Studio:

```powershell
.\gradlew --stop
```

Then, if needed, delete generated build folders manually:

```powershell
Remove-Item -Recurse -Force .\app\build -ErrorAction SilentlyContinue
Remove-Item -Recurse -Force .\build -ErrorAction SilentlyContinue
Remove-Item -Recurse -Force .\.gradle -ErrorAction SilentlyContinue
```

Then build again:

```powershell
.\gradlew assembleDebug
```

---

## Testing with the Python client

The existing Python script can still be used to verify that the broker/backend path works.

After regenerating certificates, make sure the Python script uses the current files:

```text
secrets/ca.crt
secrets/web.crt
secrets/web.key
```

If `send_image.py` starts failing with certificate errors after regenerating certificates, it usually means that one component is still using old certificate material.

The broker, Python script, and Android app must all use certificates from the same current CA generation.

---

## Troubleshooting

### `Connection refused`

The broker is not reachable.

Check that the Docker stack is running:

```bash
docker ps
```

Also check that the broker exposes port `8883`.

The Android app expects:

```text
ssl://10.0.2.2:8883
```

---

### `No subject alternative names matching IP address 10.0.2.2`

The broker certificate does not contain the emulator host address.

Fix:

1. Add `10.0.2.2` to the SAN list in the certificate generation script.
2. Regenerate certificates.
3. Restart the broker.
4. Rebuild/reinstall the Android app if local TLS files changed.

---

### `certificate verify failed`

This usually means that the CA/certificate files are inconsistent.

Common causes:

- broker is using an old `server.crt`
- Android has an old `ca.crt`
- Android has an old `android_client.p12`
- Python script uses an old `ca.crt`
- certificates were regenerated, but the broker was not restarted
- certificates were regenerated, but the Android `.p12` was not recreated

Fix by regenerating and recopying all required material:

```bash
./scripts/gen-certs.sh

openssl pkcs12 -export \
  -in secrets/web.crt \
  -inkey secrets/web.key \
  -certfile secrets/ca.crt \
  -out android_client.p12 \
  -name android-client \
  -passout pass:changeit

cp android_client.p12 android-client/app/src/main/res/raw/android_client.p12
cp secrets/ca.crt android-client/app/src/main/res/raw/ca.crt

docker compose restart broker
```

Then rebuild/reinstall the Android app.

---

### `keystore password was incorrect`

The password used to create `android_client.p12` does not match the password in `MqttUploader.kt`.

Check:

```kotlin
private const val P12_PASSWORD = "changeit"
```

If the `.p12` was generated with another password, update the Kotlin constant or regenerate the `.p12` with `changeit`.

---

### Broker log shows `tls alert certificate unknown`

This usually means that the broker rejected the client certificate.

Common causes:

- Android did not send a client certificate
- `android_client.p12` does not contain the private key
- `android_client.p12` was generated from old certificates
- broker trusts a different CA than the one used to sign the Android client certificate

Verify the `.p12` file:

```bash
openssl pkcs12 -in android_client.p12 -info -nodes -passin pass:changeit
```

The output should include both:

```text
Certificate bag
Private Key bag
```

---

## Security notes

Do not commit private key material.

The following files must stay local:

```text
android_client.p12
*.p12
*.key
```

The Android `.p12` file contains a client private key and must not be committed.

The CA private key must also never be committed or copied into the Android app:

```text
secrets/ca.key
```

The Android project should ignore local TLS runtime files:

```gitignore
android-client/app/src/main/res/raw/android_client.p12
android-client/app/src/main/res/raw/*.p12
android-client/app/src/main/res/raw/*.key
android-client/app/src/main/res/raw/ca.crt
```

---

## Current limitations

The current Android client is intentionally simple.

Current limitations:

- upload queue is displayed in memory
- full persistent offline retry is not yet implemented
- WorkManager retry can be added in a follow-up step
- Room database persistence can be added in a follow-up step
- device ID is currently hardcoded as `android-emulator-1`
- broker URL is currently hardcoded as `ssl://10.0.2.2:8883`

---

## Suggested follow-up work

Recommended next improvements:

1. Persist the upload queue using Room.
2. Add WorkManager to retry failed uploads automatically.
3. Add a settings screen for broker URL and device ID.
4. Improve user-facing error messages.
5. Add basic upload logs for demo/debugging.
6. Add a small test image fixture for local validation.

---

## Git notes

Before committing, verify that TLS material is not staged:

```bash
git status --short
```

If `android_client.p12` or `ca.crt` were accidentally staged, remove them from Git tracking while keeping the local files:

```bash
git rm --cached android-client/app/src/main/res/raw/android_client.p12
git rm --cached android-client/app/src/main/res/raw/ca.crt
```

Then commit only the Android source code, Gradle files, README, and `.gitignore` updates.