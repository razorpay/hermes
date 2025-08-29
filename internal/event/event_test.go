package event

// import (
// 	"context"
// 	"testing"

// 	"bou.ke/monkey"
// 	"github.com/golang/mock/gomock"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/zap/zapcore"

// 	"github.com/razorpay/stork/pkg/kafka"
// 	"github.com/razorpay/stork/pkg/kafka/mock"
// 	logpkg "github.com/razorpay/stork/pkg/logger"
// )

// func TestPushEvent(t *testing.T) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	logpkg.NewLogger("test", map[string]interface{}{}, zapcore.NewNopCore())

// 	ctrl := gomock.NewController(t)
// 	mp1 := mock.NewMockProducerer(ctrl)
// 	e := Event{}
// 	eventTopics = map[string]string{
// 		"message_channels.whatsapp":          "events.message_channel_whatsapp.v2.live",
// 		"webhook-delivery":                   "events.message_channel_webhook.v2.live",
// 		"message_channels.sms":               "events.message_channel_sms.v2.live",
// 		"email-delivery":                     "events.message_channel_email_delivery.v2.live",
// 		"message_channels.email":             "events.message_channel_email.v2.live",
// 		"message_channels.push_notification": "events.message_channel_push_notifications.v2.live",
// 	}
// 	ep = eventProducer{ev: &Config{
// 		Kafka:    kafka.ProducerConfig{},
// 		Disabled: true,
// 	}}
// 	e.Type = "webhook-delivery"
// 	err := PushEvent(ctx, &e)
// 	assert.Nil(t, err, "Events not enabled")

// 	ep.ev.Disabled = false
// 	e.Property = map[string]interface{}{"name": make(chan string)}
// 	err = PushEvent(ctx, &e)
// 	assert.Error(t, err, "error in converting event to JSON")

// 	e.Property = map[string]interface{}{"event_name": "webhook"}
// 	err = PushEvent(ctx, &e)
// 	assert.Error(t, err, "event producer is nil")

// 	ep.producer = mp1
// 	mp1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).Times(2)
// 	err = PushEvent(ctx, &e)
// 	assert.Nil(t, err)

// 	// https://www.calhoun.io/when-nil-isnt-equal-to-nil/
// 	var producer kafka.Producerer = nil
// 	ep.producer = producer
// 	err = PushEvent(ctx, &e)
// 	assert.Errorf(t, err, "error in pushing event into kafka")
// }

// func TestInit(t *testing.T) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	c := &Config{
// 		Kafka:    kafka.ProducerConfig{},
// 		Disabled: false,
// 	}

// 	patch := monkey.Patch(kafka.NewAsyncProducer, func(ctx context.Context, conf kafka.ProducerConfig) (kafka.Producerer, error) {
// 		return &kafka.AsyncProducer{}, nil
// 	})
// 	defer patch.Unpatch()

// 	err := Init(ctx, c)
// 	assert.Nil(t, err)
// }
