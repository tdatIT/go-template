package mqtt

import (
	"context"
	"fmt"
	"log/slog"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/tdatIT/go-template/config"
)

// MessageHandler is called for every message received on a subscribed topic.
type MessageHandler func(topic string, payload []byte)

// Client wraps a paho connection with a simplified interface.
type Client interface {
	Publish(topic string, qos byte, retained bool, payload any) error
	Subscribe(topic string, qos byte, handler MessageHandler) error
	Disconnect()
}

type mqttClient struct {
	client pahomqtt.Client
}

// NewMQTTClient creates and connects an MQTT client using the given config.
// Returns an error if the broker is unreachable within ConnectTimeout.
func NewMQTTClient(cfg *config.AppConfig) (Client, error) {
	opts := pahomqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTT.Broker)
	opts.SetClientID(cfg.MQTT.ClientID)
	opts.SetUsername(cfg.MQTT.Username)
	opts.SetPassword(cfg.MQTT.Password)
	opts.SetConnectTimeout(cfg.MQTT.ConnectTimeout)
	opts.SetKeepAlive(cfg.MQTT.KeepAlive)
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(func(_ pahomqtt.Client) {
		slog.Info("mqtt client connected", slog.String("broker", cfg.MQTT.Broker))
	})
	opts.SetConnectionLostHandler(func(_ pahomqtt.Client, err error) {
		slog.Warn("mqtt connection lost", slog.String("error", err.Error()))
	})
	opts.SetReconnectingHandler(func(_ pahomqtt.Client, _ *pahomqtt.ClientOptions) {
		slog.Info("mqtt reconnecting...")
	})

	c := pahomqtt.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.MQTT.ConnectTimeout)
	defer cancel()

	token := c.Connect()
	select {
	case <-token.Done():
		if err := token.Error(); err != nil {
			return nil, fmt.Errorf("mqtt connect: %w", err)
		}
	case <-ctx.Done():
		return nil, fmt.Errorf("mqtt connect: timeout after %s", cfg.MQTT.ConnectTimeout)
	}

	return &mqttClient{client: c}, nil
}

func (m *mqttClient) Publish(topic string, qos byte, retained bool, payload any) error {
	token := m.client.Publish(topic, qos, retained, payload)
	token.Wait()
	return token.Error()
}

func (m *mqttClient) Subscribe(topic string, qos byte, handler MessageHandler) error {
	token := m.client.Subscribe(topic, qos, func(_ pahomqtt.Client, msg pahomqtt.Message) {
		handler(msg.Topic(), msg.Payload())
	})
	token.Wait()
	return token.Error()
}

func (m *mqttClient) Disconnect() {
	m.client.Disconnect(500)
}
