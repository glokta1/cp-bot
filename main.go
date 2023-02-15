package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	// "os/signal"
	// "syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"

	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"github.com/gocolly/colly"
)


type CodeforcesResponse struct {
	Status string `json:"status"`
	Result []struct {
		ID                  int    `json:"id"`
		Name                string `json:"name"`
		Type                string `json:"type"`
		Phase               string `json:"phase"`
		DurationSeconds     int    `json:"durationSeconds"`
		StartTimeSeconds    int    `json:"startTimeSeconds"`
	} `json:"result"`
}

type JID struct {
	User   string
	Agent  uint8
	Device uint8
	Server string
	AD     bool
}

type Contest struct {
	Name string
	Link string
	Duration string
	startTime int
}

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		fmt.Println("Received a message!", v.Message.GetConversation())
	}
}

func setupWAClient() *whatsmeow.Client {
	container, err := sqlstore.New("sqlite3", "file:wastore.db?_foreign_keys=on", nil)
	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	return client
}
func main() {
	getCFContests()

	client := setupWAClient()

	defer client.Disconnect()


	var contests []Contest

	jid_list := []string{"918795927706-1634104315@g.us"}
	for _, jid_raw := range jid_list {
		jid, err := types.ParseJID(jid_raw)
		if err != nil {
			fmt.Println("Couldn't parse jid string")
		}

		msgtext := "this better work"
		msg := &proto.Message{Conversation: &msgtext,
		}
		client.SendMessage(context.Background(), jid, msg)
	}
	
}

func getCFContests() {
	resp, err := http.Get("https://codeforces.com/api/contest.list")
	if err != nil {
		fmt.Println("Couldn't get response from Codeforces API")
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)

	err = json.Unmarshal(respBytes, &contests)
	if (err != nil) {
		fmt.Println(err)
	}

}


