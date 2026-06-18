package mqttpub

import (
	"fmt"
	"log/slog"

	"github.com/bytedance/sonic"

	mqttclient "github.com/tdatIT/go-template/pkgs/mqtt"
)

type publisher struct {
	client mqttclient.Client
}

func NewPublisher(client mqttclient.Client) Publisher {
	return &publisher{client: client}
}

func (p *publisher) Publish(topic string, qos byte, payload any) error {
	if payload == nil {
		return nil
	}

	bytes, err := sonic.Marshal(payload)
	if err != nil {
		return err
	}

	if err := p.client.Publish(topic, qos, false, bytes); err != nil {
		slog.Error("mqtt publish failed", slog.String("topic", topic), slog.String("error", err.Error()))
		return fmt.Errorf("mqtt publish to %s: %w", topic, err)
	}

	return nil
}

func (p *publisher) convertPayload(payload any) ([]byte, error) {
	data, err := sonic.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal payload", slog.String("error", err.Error()))
		return nil, err
	}

	return data, nil
}
