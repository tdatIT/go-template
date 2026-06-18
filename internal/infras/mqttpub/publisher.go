package mqttpub

// Publisher sends messages to an MQTT broker.
type Publisher interface {
	Publish(topic string, qos byte, payload any) error
}
