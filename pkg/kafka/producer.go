package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"os"
	"time"

	"github.com/Shopify/sarama"

	"github.com/razorpay/goutils/kafka"
	"github.com/razorpay/stork/pkg/logger"
)

// RandomPartitioner is an input for producer config
const RandomPartitioner = "random"

var compressionCodecMap = map[string]sarama.CompressionCodec{
	"none":   sarama.CompressionNone,
	"gzip":   sarama.CompressionGZIP,
	"snappy": sarama.CompressionSnappy,
	"lz4":    sarama.CompressionLZ4,
	"zstd":   sarama.CompressionZSTD,
}

var partitionerMap = map[string]sarama.PartitionerConstructor{
	"random":     sarama.NewRandomPartitioner,
	"roundrobin": sarama.NewRoundRobinPartitioner,
	"hash":       sarama.NewHashPartitioner,
}

// Producerer is an interface used by KafkaProducer.
type Producerer interface {
	Send(ctx context.Context, msg string, topic string) error
	Disconnect(ctx context.Context)
}

// AsyncProducer : States config for the async producer.
type AsyncProducer struct {
	*kafka.KafkaProducerQueue
}

// NewAsyncProducer : Creates a new Kafka Async Producer using goutils library.
func NewAsyncProducer(ctx context.Context, conf ProducerConfig) (Producerer, error) {
	c := kafka.ProducerConfig{
		RetryBackoff:    conf.RetryBackoff,
		MaxRetry:        conf.MaxRetry,
		MaxMessages:     conf.MaxMessages,
		Brokers:         conf.Brokers,
		EnableTLS:       conf.EnableTLS,
		UserCertificate: conf.UserCertificate,
		UserKey:         conf.UserKey,
		CACertificate:   conf.CACertificate,
		DebugEnabled:    conf.DebugEnabled,
		Partitioner:     RandomPartitioner,
	}

	producer, err := kafka.NewKafkaProducer(ctx, &c)
	if err != nil {
		return nil, err
	}

	p := &AsyncProducer{producer}
	go p.HandleAck(ctx)

	return p, nil
}

// Send : send message to topic
func (p *AsyncProducer) Send(ctx context.Context, msg string, topic string) error {
	return p.SendMessage(ctx, msg, topic)
}

// Disconnect : disconnect the kafka producer
func (p *AsyncProducer) Disconnect(ctx context.Context) {
	logger.Ctx(ctx).Infow("kafka producer disconnected")
	p.KafkaProducerQueue.Disconnect()
}

// HandleAck : handle send message acknowledgement
func (p *AsyncProducer) HandleAck(ctx context.Context) {
FOR:
	for {
		select {
		case err := <-p.Producer.Errors():
			if err != nil {
				logger.Ctx(ctx).Errorw("kafka failed to send message", "error_reason", err, "error_message", err.Msg)
			}

		case <-p.Producer.Successes():

		case <-ctx.Done():
			logger.Ctx(ctx).Infow("Shutdown initiated")
			break FOR
		}
	}
	logger.Ctx(ctx).Infow("Shutdown ended")
}

// SyncProducer : States config for the sync producer.
type SyncProducer struct {
	pq *ProducerQueue
}

// ProducerQueue ...
type ProducerQueue struct {
	Producer       sarama.SyncProducer
	ctx            context.Context
	cancelFunction context.CancelFunc
	config         *sarama.Config
}

// NewSyncProducer : Creates a new Kafka sync Producer using Sarama package.
func NewSyncProducer(ctx context.Context, conf ProducerConfig) (Producerer, error) {
	c := kafka.ProducerConfig{
		MaxRetry:        conf.MaxRetry,
		MaxMessages:     conf.MaxMessages,
		Brokers:         conf.Brokers,
		EnableTLS:       conf.EnableTLS,
		UserCertificate: conf.UserCertificate,
		UserKey:         conf.UserKey,
		CACertificate:   conf.CACertificate,
		DebugEnabled:    conf.DebugEnabled,
		Partitioner:     RandomPartitioner,
	}

	producer, err := newKafkaProducer(ctx, &c)
	if err != nil {
		return nil, err
	}

	p := &SyncProducer{producer}

	return p, nil
}

func newKafkaProducer(ctx context.Context, pConfig *kafka.ProducerConfig) (*ProducerQueue, error) {
	config, err := getProducerConfig(pConfig)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		return nil, errors.New("empty Context in constructor")
	}

	producer, err := sarama.NewSyncProducer(pConfig.Brokers, config)
	if err != nil {
		return nil, err
	}
	// Rouge context check. We will need some context for graceful termination
	if ctx == nil {
		ctx = context.Background()
	}
	ctxCancel, cancelFunction := context.WithCancel(ctx)
	return &ProducerQueue{Producer: producer, ctx: ctxCancel, cancelFunction: cancelFunction, config: config}, nil
}

func getProducerConfig(pConfig *kafka.ProducerConfig) (*sarama.Config, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	kafkaVersion := pConfig.KafkaVersion
	if kafkaVersion == "" {
		kafkaVersion = kafka.DefaultKafkaVersion
	}
	version, err := sarama.ParseKafkaVersion(kafkaVersion)
	if err != nil {
		return nil, err
	}
	config.Version = version
	if isTimeFieldSet(pConfig.RetryBackoff) {
		config.Metadata.Retry.Backoff = pConfig.RetryBackoff
		config.Producer.Retry.Backoff = pConfig.RetryBackoff
	} else {
		config.Metadata.Retry.Backoff = kafka.DefaultMetadataRetryBackoff
		config.Producer.Retry.Backoff = kafka.DefaultProducerBackoff
	}
	if isIntFieldSet(pConfig.MaxMessages) {
		config.Producer.Flush.MaxMessages = pConfig.MaxMessages
	} else {
		config.Producer.Flush.MaxMessages = kafka.DefaultMaxFlushMessages
	}
	if isIntFieldSet(pConfig.MaxRetry) {
		config.Producer.Retry.Max = pConfig.MaxRetry
	} else {
		config.Producer.Retry.Max = kafka.DefaultMaxProducerRetry
	}

	//Authentication + TLS
	if pConfig.EnableTLS {
		if pConfig.CACertificate != "" && pConfig.UserCertificate != "" && pConfig.UserKey != "" {
			tlsConfig, err := newTLSConfig(pConfig.UserCertificate, pConfig.UserKey, pConfig.CACertificate)
			if err != nil {
				return nil, err
			} else {
				config.Net.TLS.Enable = true
				config.Net.TLS.Config = tlsConfig
			}
		} else {
			err := errors.New("TLS Enabled but one of the required fields of cacert/usercert/userkey is empty. Avoiding TLS")
			return nil, err
		}
	}

	// Compression
	if pConfig.CompressionEnabled {
		if pConfig.CompressionType != "none" {
			if val, ok := compressionCodecMap[pConfig.CompressionType]; ok {
				config.Producer.Compression = val
			} else {
				config.Producer.Compression = kafka.DefaultCompression
			}
		}
	} else {
		config.Producer.Compression = sarama.CompressionNone
	}

	//Partitioning
	if val, ok := partitionerMap[pConfig.Partitioner]; ok {
		config.Producer.Partitioner = val
	} else {
		config.Producer.Partitioner = sarama.NewRandomPartitioner
	}
	if pConfig.DebugEnabled {
		// Sarama quite stupidly doesn't allow any other logger other than standard logger. So, in some stupid ways, we are extending this
		sarama.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	// Check Ack Type
	config.Producer.RequiredAcks = getSupportedRetryAcks(pConfig.RetryAck)
	// Initialize other constants
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.ChannelBufferSize = 256
	config.Net.ReadTimeout = 2 * time.Second

	return config, nil
}

func isTimeFieldSet(v time.Duration) bool {
	return v > 0
}

func isIntFieldSet(v int) bool {
	return v > 0
}

// Ref: https://github.com/Shopify/sarama/blob/master/examples/sasl_scram_client/main.go#L37
func newTLSConfig(userCert, userKey, caCert string) (*tls.Config, error) {
	tlsConfig := tls.Config{}

	// Load client cert
	cert, err := tls.X509KeyPair([]byte(userCert), []byte(userKey))
	if err != nil {
		return &tlsConfig, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	if len(caCert) <= 0 {
		return &tlsConfig, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(caCert))
	tlsConfig.RootCAs = caCertPool
	tlsConfig.InsecureSkipVerify = true
	tlsConfig.Certificates = []tls.Certificate{cert}
	return &tlsConfig, err
}

func getSupportedRetryAcks(ackType string) sarama.RequiredAcks {
	switch ackType {
	case "waitforall":
		return sarama.WaitForAll
	case "noresponse":
		return sarama.NoResponse
	case "waitforlocal":
		return sarama.WaitForLocal
	case "default":
		return sarama.WaitForAll
	}
	//golang schenanigans :(
	return sarama.WaitForAll
}

// Send : send message to topic
func (p *SyncProducer) Send(ctx context.Context, message string, topic string) error {
	// publish sync
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
	_, _, err := p.pq.Producer.SendMessage(msg)
	if err != nil {
		return err
	}

	return nil
}

// Disconnect : disconnect the kafka producer.
func (p *SyncProducer) Disconnect(ctx context.Context) {
	logger.Ctx(ctx).Infow("kafka producer disconnected")
	p.pq.Disconnect()
}

// Disconnect ...
func (q ProducerQueue) Disconnect() {
	q.cancelFunction()
	err := q.Producer.Close()
	if err != nil {
		return
	}
}
