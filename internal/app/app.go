package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"aggregator/internal/controller"
	"aggregator/internal/gateway"

	"nhooyr.io/websocket"
)

type Server struct {
	Logf           func(f string, v ...interface{})
	EventPublisher *controller.EventPublisher
	SocketGateway  *gateway.SocketGateway
}

type Event struct {
	Type    string `json:"type"`
	Message any    `json:"message"`
}

type Message struct {
	Type           int     `json:"type"`
	Number         int     `json:"number"`
	Status         int     `json:"status"`
	Charge         int     `json:"charge"`
	Temperature_MK int     `json:"temperature_MK"`
	Data           DataMsg `json:"data"`
}

type MessageOne struct {
	Message
	Controlerleack ControlerLeack `json:"controlerleack"`
}

type MessageTwo struct {
	Message
	Controlermodule ControlerModule `json:"controlermodule"`
}

type MessageThree struct {
	Message
	Controlerenviroment ControlerEnviroment `json:"controlerenviroment"`
}

type DataMsg struct {
	Second int `json:"second"`
	Minute int `json:"minute"`
	Hour   int `json:"hour"`
	Day    int `json:"day"`
	Month  int `json:"month"`
	Year   int `json:"year"`
}

type ControlerLeack struct {
	Leack int `json:"leack"`
}

type ControlerModule struct {
	Temperature int `json:"temperature"`
	Humidity    int `json:"humidity"`
	Pressure    int `json:"pressure"`
	Gas         int `json:"gas"`
}

type ControlerEnviroment struct {
	Temperature int `json:"temperature"`
	Humidity    int `json:"humidity"`
	Pressure    int `json:"pressure"`
	Voc         int `json:"voc"`
	Gas1        int `json:"gas1"`
	Gas2        int `json:"gas2"`
	Gas3        int `json:"gas3"`
	Pm1         int `json:"pm1"`
	Pm25        int `json:"pm25"`
	Pm10        int `json:"pm10"`
	Fire        int `json:"fire"`
	Smoke       int `json:"smoke"`
}

type Error struct {
	Type   string `json:"type"`
	Detail string `json:"detail"`
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var err error
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		s.Logf("%v", err)
		return
	}
	defer conn.Close(websocket.StatusInternalError, "the sky is falling")

	typecontroler, err := strconv.Atoi(r.Header.Get("typecontroler"))
	if err != nil || typecontroler == 0 {
		errMsg, _ := json.Marshal(Error{"error", "typecontroler in header http/https request not send"})
		conn.Close(websocket.StatusUnsupportedData, string(errMsg))
		log.Println(string(errMsg))
		return
	}
	var controllerid string
	switch typecontroler {
	case 1:
		return
	case 10:
		_, reader, err := conn.Reader(ctx)
		if err != nil {
			errMsg, _ := json.Marshal(Error{"error", err.Error()})
			conn.Close(websocket.StatusUnsupportedData, string(errMsg))
			log.Println(string(errMsg))
			return
		}
		var msg MessageOne
		err = json.NewDecoder(reader).Decode(&msg)
		if err != nil {
			errMsg, _ := json.Marshal(Error{"error", err.Error()})
			conn.Close(websocket.StatusUnsupportedData, string(errMsg))
			log.Println(string(errMsg))
			return
		}
		controllerid, err = getControlerID(typecontroler, msg.Number)
		if err != nil {
			errMsg, _ := json.Marshal(Error{"error", err.Error()})
			conn.Close(websocket.StatusUnsupportedData, string(errMsg))
			log.Println(string(errMsg))
			return
		}
		if ok, err := s.SocketGateway.IsConnected(ctx, controllerid); ok || err != nil {
			errMsg, _ := json.Marshal(Error{"error", err.Error()})
			conn.Close(websocket.StatusUnsupportedData, string(errMsg))
			log.Println(string(errMsg))
			return
		}
		s.SocketGateway.SetConnected(ctx, controllerid)
		for {
			err = processEvent(r.Context(), conn, s.EventPublisher, controllerid)
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				s.Logf("Connection close %v", r.RemoteAddr)
				s.SocketGateway.SetDisconnected(ctx, controllerid)
				jsonData, _ := json.Marshal(&Error{"disconnect", controllerid})
				s.EventPublisher.PublishMessage(jsonData, controllerid)
				return
			}
			if err != nil {
				s.Logf("Connection failed %v: %v", r.RemoteAddr, err)
				s.SocketGateway.SetDisconnected(ctx, controllerid)
				jsonData, _ := json.Marshal(&Error{"disconnect", controllerid})
				s.EventPublisher.PublishMessage(jsonData, controllerid)
				return
			}
		}
	case 11:
		_, reader, err := conn.Reader(ctx)
		if err != nil {
			errMsg, _ := json.Marshal(Error{"error", err.Error()})
			conn.Close(websocket.StatusUnsupportedData, string(errMsg))
			log.Println(string(errMsg))
			return
		}
		var msg MessageTwo
		err = json.NewDecoder(reader).Decode(&msg)
		if err != nil {
			errMsg, _ := json.Marshal(Error{"error", err.Error()})
			conn.Close(websocket.StatusUnsupportedData, string(errMsg))
			log.Println(string(errMsg))
			return
		}
		controllerid, err = getControlerID(typecontroler, msg.Number)
		if err != nil {
			errMsg, _ := json.Marshal(Error{"error", err.Error()})
			conn.Close(websocket.StatusUnsupportedData, string(errMsg))
			log.Println(string(errMsg))
			return
		}
		if ok, err := s.SocketGateway.IsConnected(ctx, controllerid); ok || err != nil {
			errMsg, _ := json.Marshal(Error{"error", err.Error()})
			conn.Close(websocket.StatusUnsupportedData, string(errMsg))
			log.Println(string(errMsg))
			return
		}
		s.SocketGateway.SetConnected(ctx, controllerid)
		for {
			err = processEvent(r.Context(), conn, s.EventPublisher, controllerid)
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				s.Logf("Connection close %v", r.RemoteAddr)
				s.SocketGateway.SetDisconnected(ctx, controllerid)
				jsonData, _ := json.Marshal(&Error{"disconnect", controllerid})
				s.EventPublisher.PublishMessage(jsonData, controllerid)
				return
			}
			if err != nil {
				s.Logf("Connection failed %v: %v", r.RemoteAddr, err)
				s.SocketGateway.SetDisconnected(ctx, controllerid)
				jsonData, _ := json.Marshal(&Error{"disconnect", controllerid})
				s.EventPublisher.PublishMessage(jsonData, controllerid)
				return
			}
		}
	case 12:
		_, reader, err := conn.Reader(ctx)
		if err != nil {
			errMsg, _ := json.Marshal(Error{"error", err.Error()})
			conn.Close(websocket.StatusUnsupportedData, string(errMsg))
			log.Println(string(errMsg))
			return
		}
		var msg MessageThree
		err = json.NewDecoder(reader).Decode(&msg)
		if err != nil {
			errMsg, _ := json.Marshal(Error{"error", err.Error()})
			conn.Close(websocket.StatusUnsupportedData, string(errMsg))
			log.Println(string(errMsg))
			return
		}
		controllerid, err = getControlerID(typecontroler, msg.Number)
		if err != nil {
			errMsg, _ := json.Marshal(Error{"error", err.Error()})
			conn.Close(websocket.StatusUnsupportedData, string(errMsg))
			log.Println(string(errMsg))
			return
		}
		if ok, err := s.SocketGateway.IsConnected(ctx, controllerid); ok || err != nil {
			errMsg, _ := json.Marshal(Error{"error", err.Error()})
			conn.Close(websocket.StatusUnsupportedData, string(errMsg))
			log.Println(string(errMsg))
			return
		}
		s.SocketGateway.SetConnected(ctx, controllerid)
		for {
			err = processEvent(r.Context(), conn, s.EventPublisher, controllerid)
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				s.Logf("Connection close %v", r.RemoteAddr)
				s.SocketGateway.SetDisconnected(ctx, controllerid)
				jsonData, _ := json.Marshal(&Error{"disconnect", controllerid})
				s.EventPublisher.PublishMessage(jsonData, controllerid)
				return
			}
			if err != nil {
				s.Logf("Connection failed %v: %v", r.RemoteAddr, err)
				s.SocketGateway.SetDisconnected(ctx, controllerid)
				jsonData, _ := json.Marshal(&Error{"disconnect", controllerid})
				s.EventPublisher.PublishMessage(jsonData, controllerid)
				return
			}
		}
	default:
		errMsg, _ := json.Marshal(Error{"error", "typecontroler in header http/https request not send"})
		conn.Close(websocket.StatusUnsupportedData, string(errMsg))
		log.Println(string(errMsg))
		return
	}
}

func processEvent(
	ctx context.Context,
	c *websocket.Conn,
	ep *controller.EventPublisher,
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

func getControlerID(typecontroler int, number int) (string, error) {
	config := GetConfig()
	apiURL := fmt.Sprintf("%s/controllers/client?controlerid=%s", config.API.URL(), strconv.Itoa(typecontroler)+"_"+strconv.Itoa(number))
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

	controleridstring, ok := payload["controlerid"]
	if !ok {
		return "", fmt.Errorf("controlerid not registered")
	}
	return controleridstring, nil
}
