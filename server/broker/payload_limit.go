package broker

import (
	"fmt"
	"os"
	"strconv"
)

const (
	DefaultMaxMQTTPayloadBytes = 5 * 1024 * 1024 // 5 MiB
	maxMQTTPayloadBytesEnv     = "MQTT_MAX_PAYLOAD_BYTES"
)

type PayloadSizeCheck struct {
	Allowed     bool
	ActualBytes int
	MaxBytes    int
}

func MaxMQTTPayloadBytes() int {
	rawValue := os.Getenv(maxMQTTPayloadBytesEnv)
	if rawValue == "" {
		return DefaultMaxMQTTPayloadBytes
	}

	parsedValue, err := strconv.Atoi(rawValue)
	if err != nil || parsedValue <= 0 {
		return DefaultMaxMQTTPayloadBytes
	}

	return parsedValue
}

func CheckMQTTPayloadSize(payload []byte) PayloadSizeCheck {
	maxBytes := MaxMQTTPayloadBytes()
	actualBytes := len(payload)

	return PayloadSizeCheck{
		Allowed:     actualBytes <= maxBytes,
		ActualBytes: actualBytes,
		MaxBytes:    maxBytes,
	}
}

func RejectOversizedMQTTPayload(topic string, payload []byte) bool {
	check := CheckMQTTPayloadSize(payload)

	if check.Allowed {
		return false
	}

	fmt.Printf(
		"SECURITY_EVENT=MQTT_PAYLOAD_SIZE_LIMIT_EXCEEDED topic=%s actual_bytes=%d max_bytes=%d\n",
		topic,
		check.ActualBytes,
		check.MaxBytes,
	)

	return true
}
