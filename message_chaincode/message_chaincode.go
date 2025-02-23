package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// 🔥 Mesaj Modeli
type Message struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
	Content        string `json:"content"`
	Timestamp      string `json:"timestamp"`
}

// 🔥 Chaincode (Akıllı Sözleşme)
type MessageContract struct {
	contractapi.Contract
}

// ✅ 1. Mesaj Ekleme Fonksiyonu
func (m *MessageContract) CreateMessage(ctx contractapi.TransactionContextInterface, id string, conversationID string, senderID string, content string, timestamp string) error {
	message := Message{
		ID:             id,
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		Timestamp:      timestamp,
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("Mesaj JSON'a çevrilemedi: %s", err.Error())
	}

	// 🔥 Blockchain’e mesaj ekle
	return ctx.GetStub().PutState(id, messageJSON)
}

// ✅ 2. Mesajları Listeleme Fonksiyonu
func (m *MessageContract) GetMessage(ctx contractapi.TransactionContextInterface, id string) (*Message, error) {
	messageJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("Mesaj okunamadı: %s", err.Error())
	}

	if messageJSON == nil {
		return nil, fmt.Errorf("Mesaj bulunamadı: %s", id)
	}

	var message Message
	err = json.Unmarshal(messageJSON, &message)
	if err != nil {
		return nil, fmt.Errorf("JSON dönüşümü başarısız: %s", err.Error())
	}

	return &message, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(MessageContract))
	if err != nil {
		fmt.Printf("Chaincode başlatılamadı: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Chaincode başlatılamadı: %s", err.Error())
	}
}
