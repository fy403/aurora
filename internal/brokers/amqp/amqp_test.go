package amqp_test

import (
	"aurora/internal/brokers/amqp"
	"aurora/internal/brokers/iface"
	"aurora/internal/config"
	"aurora/internal/tasks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdjustRoutingKey(t *testing.T) {
	t.Parallel()

	var (
		s      *tasks.Signature
		broker iface.Broker
	)

	t.Run("with routing and binding keys", func(t *testing.T) {
		s := &tasks.Signature{RoutingKey: "routing_key"}
		broker = amqp.New(&config.Config{
			DefaultQueue: "queue",
			AMQP: &config.AMQPConfig{
				ExchangeType: "direct",
				BindingKey:   "binding_key",
			},
		})
		broker.AdjustRoutingKey(s)
		assert.Equal(t, "routing_key", s.RoutingKey)
	})

	t.Run("with binding key", func(t *testing.T) {
		s = new(tasks.Signature)
		broker = amqp.New(&config.Config{
			DefaultQueue: "queue",
			AMQP: &config.AMQPConfig{
				ExchangeType: "direct",
				BindingKey:   "binding_key",
			},
		})
		broker.AdjustRoutingKey(s)
		assert.Equal(t, "binding_key", s.RoutingKey)
	})
}
