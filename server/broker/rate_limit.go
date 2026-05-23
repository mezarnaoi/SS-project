package broker

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	DefaultMQTTRateLimitMessages = 60

	DefaultMQTTRateLimitWindowSeconds = 60

	mqttRateLimitMessagesEnv      = "MQTT_RATE_LIMIT_MESSAGES"
	mqttRateLimitWindowSecondsEnv = "MQTT_RATE_LIMIT_WINDOW_SECONDS"
)

type RateLimitCheck struct {
	Allowed       bool
	MessageCount  int
	MaxMessages   int
	WindowSeconds int
}

type RateLimiter struct {
	mu         sync.Mutex
	timestamps map[string][]time.Time
}

var defaultRateLimiter = NewRateLimiter()

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		timestamps: make(map[string][]time.Time),
	}
}

func MQTTRateLimitMessages() int {
	raw := os.Getenv(mqttRateLimitMessagesEnv)
	if raw == "" {
		return DefaultMQTTRateLimitMessages
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return DefaultMQTTRateLimitMessages
	}
	return v
}

func MQTTRateLimitWindowSeconds() int {
	raw := os.Getenv(mqttRateLimitWindowSecondsEnv)
	if raw == "" {
		return DefaultMQTTRateLimitWindowSeconds
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return DefaultMQTTRateLimitWindowSeconds
	}
	return v
}

func (r *RateLimiter) Check(clientKey string) RateLimitCheck {
	maxMessages := MQTTRateLimitMessages()
	windowSeconds := MQTTRateLimitWindowSeconds()
	window := time.Duration(windowSeconds) * time.Second
	now := time.Now()
	cutoff := now.Add(-window)

	r.mu.Lock()
	defer r.mu.Unlock()

	existing := r.timestamps[clientKey]
	pruned := existing[:0]
	for _, ts := range existing {
		if ts.After(cutoff) {
			pruned = append(pruned, ts)
		}
	}

	messageCount := len(pruned) + 1
	allowed := messageCount <= maxMessages

	if allowed {
		r.timestamps[clientKey] = append(pruned, now)
	} else {
		r.timestamps[clientKey] = pruned
	}

	return RateLimitCheck{
		Allowed:       allowed,
		MessageCount:  messageCount,
		MaxMessages:   maxMessages,
		WindowSeconds: windowSeconds,
	}
}

func CheckMQTTRateLimit(clientKey string) RateLimitCheck {
	return defaultRateLimiter.Check(clientKey)
}

func RejectRateLimitedMQTTMessage(topic string) bool {
	check := CheckMQTTRateLimit(topic)
	if check.Allowed {
		return false
	}

	fmt.Printf(
		"SECURITY_EVENT=MQTT_RATE_LIMIT_EXCEEDED topic=%s count=%d max=%d window_seconds=%d\n",
		topic,
		check.MessageCount,
		check.MaxMessages,
		check.WindowSeconds,
	)

	return true
}
