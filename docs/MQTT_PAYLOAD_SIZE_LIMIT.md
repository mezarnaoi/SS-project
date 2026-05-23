# MQTT Payload Size Limit

## Security control

The application enforces a maximum MQTT payload size before processing incoming MQTT messages.

Default value:

```text
5242880 bytes -> 5 MiB
```


## Configuration

The backend limit can be configured with:

MQTT_MAX_PAYLOAD_BYTES

The broker limit is configured in:

```text
broker/mosquitto.conf
```
using:

```text
message_size_limit 5242880
```
## Security

The system receives medical images and documents through MQTT. Without a payload size limit, an attacker or malfunctioning client could send very large messages that consume memory, CPU, disk, OCR processing time, or database resources.

The backend now rejects oversized MQTT payloads before:

image decoding
OCR processing
medical certificate parsing
MongoDB writes
local file saving
Rejection behavior

Oversized messages are rejected and logged with:

```text
SECURITY_EVENT=MQTT_PAYLOAD_SIZE_LIMIT_EXCEEDED
```

## Validation

The control is covered by unit tests for:

payload below the configured limit
payload exactly equal to the configured limit
payload above the configured limit
invalid environment variable fallback
negative environment variable fallback

Run tests with:

```text
cd server
go test ./...
```