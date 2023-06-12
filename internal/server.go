package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"homepp/aggregator/internal/config"

	"nhooyr.io/websocket"
)

type Server struct {
	Logf           func(f string, v ...interface{})
	EventPublisher *EventPublisher
	SocketGateway  *SocketGateway
}

type Event struct {
	Type    string `json:"type"`
	Message any    `json:"message"`
}

type Error struct {
	Type   string `json:"type"`
	Detail string `json:"detail"`
}

func (e Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		e.Logf("%v", err)
		return
	}
	defer conn.Close(websocket.StatusInternalError, "the sky is falling")

	hwKey := r.Header.Get("Hardwarekey")
	if hwKey == "" {
		errMsg, _ := json.Marshal(Error{"error", "hardware key not not send"})
		conn.Write(ctx, 1, errMsg)
		log.Println(string(errMsg))
		return
	}

	clientID, err := getClientID(hwKey)
	if err != nil {
		errMsg, _ := json.Marshal(Error{"error", err.Error()})
		conn.Write(ctx, 1, errMsg)
		log.Println(string(errMsg))
		return
	}

	e.SocketGateway.SetConnected(ctx, hwKey)
	for {
		err = processEvent(r.Context(), conn, e.EventPublisher, clientID)
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
			return
		}
		if err != nil {
			e.Logf("Connection failed %v: %v", r.RemoteAddr, err)
			e.SocketGateway.SetDisconnected(ctx, hwKey)
			jsonData, _ := json.Marshal(&Error{"disconnect", hwKey})
			e.EventPublisher.PublishMessage(jsonData, clientID)
			return
		}
	}
}

func processEvent(
	ctx context.Context,
	c *websocket.Conn,
	ep *EventPublisher,
	clientID string,
) error {
	msgType, reader, err := c.Reader(ctx)
	if err != nil {
		return err
	}

	var event Event
	err = json.NewDecoder(reader).Decode(&event)
	if err != nil {
		jsonData, _ := json.Marshal(&Error{"error", "invalid message format"})
		c.Write(ctx, msgType, jsonData)
	}

	jsonData, _ := json.Marshal(event)
	ep.PublishMessage(jsonData, clientID)

	return err
}

func getClientID(hwKey string) (string, error) {
	config := config.GetConfig()
	apiURL := fmt.Sprintf("%s/controllers/client?hw_key=%s", config.API.URL(), hwKey)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var payload map[string]string
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	clientID, ok := payload["client_id"]
	if !ok {
		return "", fmt.Errorf("hardware key not registered")
	}

	return clientID, nil
}
