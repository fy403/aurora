package amqp

import (
	"aurora/internal/brokers/iface"
	"aurora/internal/log"
	"fmt"

	"github.com/streadway/amqp"
)

func (b *Broker) CreateSpecQueue(queueName string, consumerTag string, concurrency int, taskProcessor iface.TaskProcessor) (bool, error) {
	conn, channel, queue, _, amqpCloseChan, err := b.Connect(
		b.GetConfig().Broker,
		b.GetConfig().MultipleBrokerSeparator,
		b.GetConfig().TLSConfig,
		b.GetConfig().AMQP.Exchange,     // exchange name
		b.GetConfig().AMQP.ExchangeType, // exchange type
		queueName,                       // queue name
		true,                            // queue durable
		false,                           // queue delete when unused
		queueName,                       // queue binding key
		nil,                             // exchange declare args
		amqp.Table(b.GetConfig().AMQP.QueueDeclareArgs), // queue declare args
		amqp.Table(b.GetConfig().AMQP.QueueBindingArgs), // queue binding args
	)
	if err != nil {
		b.GetRetryFunc()(b.GetRetryStopChan())
		return b.GetRetry(), err
	}
	defer b.Close(channel, conn)

	if err = channel.Qos(
		b.GetConfig().AMQP.PrefetchCount,
		0,     // prefetch size
		false, // global
	); err != nil {
		return b.GetRetry(), fmt.Errorf("Channel qos error: %s", err)
	}

	deliveries, err := channel.Consume(
		queue.Name,  // queue
		consumerTag, // consumer tag
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return b.GetRetry(), fmt.Errorf("Queue consume error: %s", err)
	}

	log.Runtime().Infof("[spec queue: %s] Waiting for messages. To exit press CTRL+C", queue.Name)

	if err := b.consume(deliveries, concurrency, taskProcessor, amqpCloseChan); err != nil {
		return b.GetRetry(), err
	}

	// Waiting for any tasks being processed to finish
	b.processingWG.Wait()
	return b.GetRetry(), nil
}
