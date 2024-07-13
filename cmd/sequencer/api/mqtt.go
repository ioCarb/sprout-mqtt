package api

import (
	"crypto/ecdsa"
	"encoding/json"

	//"log"
	"fmt"
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

// Subscribe only to a topic for messages for now
func (s *mqttServer) Run(topic string) {
	if token := s.client.Subscribe(topic, 0, s.messageHandler); token.Wait() && token.Error() != nil {
		slog.Error("Failed to subscribe to topic", "error", token.Error())
	}
}

// TODO topic for JWK token request

// TODO implement verifyToken

func (s *mqttServer) messageHandler(client mqtt.Client, msg mqtt.Message) {
	req := &apitypes.HandleMessageReqMqtt{}
	if err := json.Unmarshal(msg.Payload(), req); err != nil {
		slog.Error("Failed to unmarshal message payload", "error", err)
		return
	}

	// Fetch client info from pool, if not found, fetch from SC
	clientInfo := s.clients.ClientByIoID(req.ClientID)
	if clientInfo == nil {
		slog.Error("Client not found", "clientID", req.ClientID)
		return
	}

	// Validate project permission
	hasPermission, err := s.clients.HasProjectPermission(clientInfo.DID(), req.ProjectID)
	if err != nil {
		slog.Error("Failed to check project permission", "error", err)
		return
	}

	if !hasPermission {
		slog.Error("Client does not have permission to project", "clientID", req.ClientID, "projectID", req.ProjectID)
		return
	}

	id := uuid.NewString()
	message := &persistence.Message{
		MessageID:      id,
		ProjectID:      req.ProjectID,
		ProjectVersion: req.ProjectVersion,
		ClientID:       req.ClientID,
		Data:           []byte(req.Data),
	}

	if err := s.p.Save(message, s.aggregationAmount, s.privateKey); err != nil {
		slog.Error("Failed to save message", "error", err)
		return
	}

	slog.Info("Message saved successfully", "messageID", id)

	response := &apitypes.HandleMessageRsp{MessageID: id}
	responseTopic := fmt.Sprintf("response/%s", req.ClientID)

	responseBytes, err := json.Marshal(response)
	if err != nil {
		slog.Error("Failed to marshal response", "error", err)
		return
	}

	if token := s.client.Publish(responseTopic, 0, false, responseBytes); token.Wait() && token.Error() != nil {
		slog.Error("Failed to publish response", "error", token.Error())
	}
}
