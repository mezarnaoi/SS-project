package broker

import (
	"testing"
)

func TestCheckMQTTPayloadSize_AllowsPayloadBelowLimit(t *testing.T) {
	t.Setenv("MQTT_MAX_PAYLOAD_BYTES", "10")

	result := CheckMQTTPayloadSize([]byte("hello"))

	if !result.Allowed {
		t.Fatalf("expected payload to be allowed")
	}

	if result.ActualBytes != 5 {
		t.Fatalf("expected actual size 5, got %d", result.ActualBytes)
	}

	if result.MaxBytes != 10 {
		t.Fatalf("expected max size 10, got %d", result.MaxBytes)
	}
}

func TestCheckMQTTPayloadSize_AllowsPayloadEqualToLimit(t *testing.T) {
	t.Setenv("MQTT_MAX_PAYLOAD_BYTES", "5")

	result := CheckMQTTPayloadSize([]byte("hello"))

	if !result.Allowed {
		t.Fatalf("expected payload equal to limit to be allowed")
	}

	if result.ActualBytes != 5 {
		t.Fatalf("expected actual size 5, got %d", result.ActualBytes)
	}

	if result.MaxBytes != 5 {
		t.Fatalf("expected max size 5, got %d", result.MaxBytes)
	}
}

func TestCheckMQTTPayloadSize_RejectsPayloadAboveLimit(t *testing.T) {
	t.Setenv("MQTT_MAX_PAYLOAD_BYTES", "4")

	result := CheckMQTTPayloadSize([]byte("hello"))

	if result.Allowed {
		t.Fatalf("expected payload above limit to be rejected")
	}

	if result.ActualBytes != 5 {
		t.Fatalf("expected actual size 5, got %d", result.ActualBytes)
	}

	if result.MaxBytes != 4 {
		t.Fatalf("expected max size 4, got %d", result.MaxBytes)
	}
}

func TestMaxMQTTPayloadBytes_UsesDefaultForInvalidValue(t *testing.T) {
	t.Setenv("MQTT_MAX_PAYLOAD_BYTES", "invalid")

	result := MaxMQTTPayloadBytes()

	if result != DefaultMaxMQTTPayloadBytes {
		t.Fatalf("expected default value %d, got %d", DefaultMaxMQTTPayloadBytes, result)
	}
}

func TestMaxMQTTPayloadBytes_UsesDefaultForNegativeValue(t *testing.T) {
	t.Setenv("MQTT_MAX_PAYLOAD_BYTES", "-1")

	result := MaxMQTTPayloadBytes()

	if result != DefaultMaxMQTTPayloadBytes {
		t.Fatalf("expected default value %d, got %d", DefaultMaxMQTTPayloadBytes, result)
	}
}
