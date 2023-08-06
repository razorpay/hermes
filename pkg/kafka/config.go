package kafka

import "time"

// ProducerConfig states config of th kafka broker
type ProducerConfig struct {
	RetryBackoff    time.Duration
	Topic           string
	MaxRetry        int
	MaxMessages     int
	Brokers         []string
	EnableTLS       bool
	UserCertificate string
	UserKey         string
	CACertificate   string
	DebugEnabled    bool
}
