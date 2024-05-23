package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Homeppv2/aggregator-go/internal/controller"
	"github.com/Homeppv2/entitys"
	"nhooyr.io/websocket"
)

type Server struct {
	Logf           func(f string, v ...interface{})
	EventPublisher *controller.EventPublisher
}

type Event struct {
	Type    string `json:"type"`
	Message any    `json:"message"`
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
	s.Logf("connected")
	s.Logf("start cheking data")
	t := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-t.C:
			_, data, err := conn.Read(ctx)
			if err != nil {
				errMsg, _ := json.Marshal(&Error{"error", err.Error()})
				log.Println(string(errMsg))
				continue
			}
			var msgZiro entitys.MessangeTypeZiroJson
			err = json.Unmarshal(data, &msgZiro)
			// нормально пофиксить
			s.Logf("%s", msgZiro)
			if err != nil {
				errMsg, _ := json.Marshal(&Error{"error", err.Error()})
				log.Println(string(errMsg))
				continue
			}

			if msgZiro.RequestAuth == nil {
				continue
			}
			var id string
			if msgZiro.One != nil {
				id, err = getControlerID(msgZiro.One.Type, msgZiro.One.Number, msgZiro.RequestAuth.Email, msgZiro.RequestAuth.Password)
			}
			if msgZiro.Two != nil {
				id, err = getControlerID(msgZiro.Two.Type, msgZiro.Two.Number, msgZiro.RequestAuth.Email, msgZiro.RequestAuth.Password)
			}
			if msgZiro.Three != nil {
				id, err = getControlerID(msgZiro.Three.Type, msgZiro.Three.Number, msgZiro.RequestAuth.Email, msgZiro.RequestAuth.Password)
			}
			if err != nil {
				errMsg, _ := json.Marshal(&Error{"error", err.Error()})
				log.Println(string(errMsg))
				continue
			}
			s.Logf("gettet id controller = %s", id)
			err = processEvent(r.Context(), msgZiro, s.EventPublisher, id)
			if err != nil {
				errMsg, _ := json.Marshal(&Error{"error", err.Error()})
				log.Println(string(errMsg))
			}
		}
	}
}

func processEvent(
	ctx context.Context,
	msg entitys.MessangeTypeZiroJson,
	ep *controller.EventPublisher,
	clientID string,
) error {
	jsonData, err := json.Marshal(msg)
	ep.PublishMessage(jsonData, clientID)
	return err
}

func getControlerID(typecontroler int, number int, email, password string) (string, error) {
	apiURL := fmt.Sprintf("%s/getidcontroller", fmt.Sprintf("%s://%s:%s", os.Getenv("API_PROTOCOL"), os.Getenv("API_HOST"), os.Getenv("API_PORT")))
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("type", strconv.Itoa(typecontroler))
	req.Header.Set("number", strconv.Itoa(number))
	req.Header.Set("email", email)
	req.Header.Set("password", password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	str := resp.Header.Get("idcontroller")
	if str == "-1" {
		return "", errors.New("403")
	}
	if str == "" {
		return "", errors.New("не получен айди")
	}
	return str, nil
}

func readData(reader io.Reader) []byte {
	ans, _ := io.ReadAll(reader)
	return ans
}
