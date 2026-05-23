# SS-project

| Branch / Feature | Ownership |
|------------------|-----------|
| Preluarea datelor & API-uri: broker MQTT | ` ` |
| Preluarea datelor & API-uri: endpoint-uri web | ` ` |
| Dark mode toogle | ` ` |
| Raportare dinamică și stocare intr-o baza de date | ` ` |
| API de generare rapoarte extensibil | ` ` |
| Izolarea Procesării Imaginilor (OCR Sandboxing) | ` ` |
| Protecția datelor din baza de date | ` ` |
| Controlul accesului la date | ` ` |
| Captură de imagini utilizând camera dispozitivului | ` ` |
| Transmisie securizată prin MQTT folosind mTLS sau similar | ` ` |
| Mod offline (stocare locală) | ` ` |


Write UID and GUID in .env
```bash
id -u && id -g
```

The Android app expects the following local files:

```text
app/src/main/res/raw/ca.crt
app/src/main/res/raw/android_client.p12
```

These files are not committed to Git because they are generated environment-specific TLS materials.

android_client.p12 contains the client certificate and private key, so it must never be committed.

## Generating certificates

From the project root, regenerate the project certificates:
## mTLS
```bash
bash scripts/gen-certs.sh
```
or use with this option if it throws an error
```bash
MSYS_NO_PATHCONV=1 bash scripts/gen-certs.sh
```
Creating the Android PKCS#12 keystore

Android uses a PKCS#12 keystore for the client certificate and private key.

## From the project root:
```bash
openssl pkcs12 -export \
  -in secrets/web.crt \
  -inkey secrets/web.key \
  -certfile secrets/ca.crt \
  -out android_client.p12 \
  -name android-client \
  -passout pass:changeit
```

Then copy the TLS files into the Android project:

cp android_client.p12 android-client/app/src/main/res/raw/android_client.p12
cp secrets/ca.crt android-client/app/src/main/res/raw/ca.crt

## Recreate the stack:

docker compose down
docker compose up -d
Running the Android app









