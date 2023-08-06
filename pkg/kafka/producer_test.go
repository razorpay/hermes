package kafka

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProducer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := ProducerConfig{}
	prod, err := NewAsyncProducer(ctx, config)
	assert.Nil(t, prod)
	assert.Error(t, err, "abc")
	cancel()
}
