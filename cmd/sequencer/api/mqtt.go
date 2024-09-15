package api

import (
	"crypto/ecdsa"
	"encoding/json"
	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"

	"github.com/machinefi/sprout/apitypes"
	"github.com/machinefi/sprout/clients"
	"github.com/machinefi/sprout/cmd/sequencer/persistence"
)

type mqttServer struct {
	client             mqtt.Client
	p                  *persistence.Persistence
	coordinatorAddress string
	aggregationAmount  uint
	privateKey         *ecdsa.PrivateKey
	clients            *clients.Manager
}

func NewMqttServer(p *persistence.Persistence, aggregationAmount uint, coordinatorAddress string, sk *ecdsa.PrivateKey, broker string, clientMgr *clients.Manager) *mqttServer {
	opts := mqtt.NewClientOptions().AddBroker(broker)
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		slog.Error("Failed to connect to MQTT broker", "error", token.Error())
	}
	return &mqttServer{
		client:             client,
		p:                  p,
		coordinatorAddress: coordinatorAddress,
		aggregationAmount:  aggregationAmount,
		privateKey:         sk,
		clients:            clientMgr,
	}
}

// Subscribe only to a topic for messages
func (s *mqttServer) Run(topic string) {
	if token := s.client.Subscribe(topic, 0, s.messageHandler); token.Wait() && token.Error() != nil {
		slog.Error("Failed to subscribe to topic", "error", token.Error())
	}
}

func (s *mqttServer) messageHandler(client mqtt.Client, msg mqtt.Message) {

	// Unmarshall payload into HandleMessageReq
	req := &apitypes.HandleMessageReq{}
	if err := json.Unmarshal(msg.Payload(), req); err != nil {
		slog.Error("Failed to unmarshal message payload", "error", err)
		return
	}

	// TODO check if client is authorized to send messages
	// TODO check project permissions

	id := uuid.NewString()
	message := &persistence.Message{
		MessageID:      id,
		ProjectID:      req.ProjectID,
		ProjectVersion: req.ProjectVersion,
		Data:           []byte(req.Data),
	}

	if err := s.p.Save(message, s.aggregationAmount, s.privateKey); err != nil {
		slog.Error("Failed to save message", "error", err)
		return
	}

	slog.Info("Message saved successfully", "messageID", id)
}
