package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/razorpay/stork/pkg/kafka"
	"github.com/razorpay/stork/pkg/logger"
)

const mode = "live"
const source = "stork"

type eventProducer struct {
	ev       *Config
	producer kafka.Producerer
}

var ep eventProducer

var eventTopics = map[string]string{
	"message_channels.whatsapp":          "events.message_channel_whatsapp.v2.live",
	"webhook-delivery":                   "events.message_channel_webhook.v2.live",
	"message_channels.sms":               "events.message_channel_sms.v2.live",
	"email-delivery":                     "events.message_channel_email_delivery.v2.live",
	"message_channels.email":             "events.message_channel_email.v2.live",
	"message_channels.push_notification": "events.message_channel_push_notifications.v2.live",
}

// Config : holds kafka config
type Config struct {
	Kafka    kafka.ProducerConfig
	Disabled bool
}

// Event struct implements the Kafka Events V1 Interface.
type Event struct {
	Name              string                 `json:"event_name"`
	Type              string                 `json:"event_type"`
	Group             string                 `json:"event_group"`
	Version           string                 `json:"version"`
	EventTimestamp    int64                  `json:"event_timestamp"`
	ProducerTimestamp int64                  `json:"producer_timestamp"`
	Source            string                 `json:"source"`
	Mode              string                 `json:"mode"`
	Property          map[string]interface{} `json:"properties"`
	Context           map[string]string      `json:"context"`
	Metadata          map[string]string      `json:"metadata"`
	ReadKey           []string               `json:"read_key"`
	WriteKey          string                 `json:"write_key"`
}

// GetEventName gives the name of event.
func (e Event) GetEventName() string {
	return e.Name
}

// GetEventType gives the type of event.
func (e Event) GetEventType() string {
	return e.Type
}

// GetEventGroup gives the group of event.
func (e Event) GetEventGroup() string {
	return e.Group
}

// GetVersion gives the version.
func (e Event) GetVersion() string {
	return e.Version
}

// GetProperties gives the property of event.
func (e Event) GetProperties() map[string]interface{} {
	return e.Property
}

// GetEventTimestamp gives the timestamp.
func (e Event) GetEventTimestamp() int64 {
	return e.EventTimestamp
}

// GetProducerTimestamp gives the producer timestamp
func (e Event) GetProducerTimestamp() int64 {
	return e.ProducerTimestamp
}

// GetSource gives the source of event.
func (e Event) GetSource() string {
	return e.Source
}

// GetMode gives the mode of event.
func (e Event) GetMode() string {
	return e.Mode
}

// GetContext gives the context of event.
func (e Event) GetContext() map[string]string {
	return e.Context
}

// Init : initialize kafka event producer
func Init(ctx context.Context, eventConfig *Config, log hclog.Logger) error {
	var (
		err      error
		producer kafka.Producerer
	)
	ep = eventProducer{ev: eventConfig}
	if ep.ev.Disabled == false {
		producer, err = kafka.NewSyncProducer(ctx, ep.ev.Kafka)
		if err != nil || producer == nil {
			log.Error("client not connected. Data-loss possibilities", "error", err, "producer: ", producer)
		} else {
			ep.producer = producer
		}
	}

	return nil
}

// PushEvent : Push event in kafka
func PushEvent(ctx context.Context, e *Event) error {
	e.Mode = mode
	e.Source = source
	e.ProducerTimestamp = time.Now().Unix()
	e.Context = map[string]string{}
	e.Metadata = map[string]string{}
	e.ReadKey = []string{}
	if ep.ev.Disabled == true {
		return nil
	}

	dataByte, err := json.Marshal(e)
	if err != nil {
		logger.Ctx(ctx).Errorw("error in converting event to JSON", "error", err)
		eventPushErrorCounter.Inc()
		return err
	}

	if _, ok := eventTopics[e.Type]; !ok {
		eventPushErrorCounter.Inc()
		logger.Ctx(ctx).Infow("event category is missing. Data loss possibilities.", "event", e)
	}

	if ep.producer == nil {
		err = fmt.Errorf("event producer is nil")
		eventPushErrorCounter.Inc()
		logger.Ctx(ctx).Errorw("error in pushing event into kafka", "error", err, "event", e)
		return err
	}
	startTime := time.Now()
	if err = ep.producer.Send(ctx, string(dataByte), eventTopics[e.Type]); err != nil {
		eventPushErrorCounter.Inc()
		logger.Ctx(ctx).Errorw("error in pushing event into kafka", "error", err, "event", e)
	}
	eventPushTimeHistMs.Observe(float64(time.Since(startTime).Milliseconds()))
	eventPushCounter.Inc()
	return nil
}

// Shutdown : for graceful shutdown
func Shutdown(ctx context.Context) {
	if ep.producer != nil {
		ep.producer.Disconnect(ctx)
	}
}
