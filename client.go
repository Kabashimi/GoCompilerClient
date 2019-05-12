package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

//var addr2 = flag.String("addr2", "localhost:8080", "http service address")
var addr2 = flag.String("addr2", "192.168.0.17:8080", "http service address")

func main() {

	file_to_send, err := ioutil.ReadFile("data/data.go")

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr2, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	awaitingResults := false
	awaitingRaport := false

	mes := ""

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			mes = string(message)
			//log.Printf("recv: %s", message)
			//awaiting first message
			if !awaitingResults {
				//compile success/failure info handle:
				if string(message) == "NOK" {
					fmt.Println("Program compilation was NOT successful")
					return
				} else if string(message) == "OK" {
					fmt.Println("Program compilation successful!")
					awaitingResults = true
				}
			} else {
				if !awaitingRaport {
					log.Println("Received results:\n", mes)
					awaitingRaport = true
				} else {
					fmt.Println("Received raport:\n", mes)
					return
				}
			}

		}
	}()

	msg := c.WriteMessage(websocket.TextMessage, []byte(string(file_to_send)))
	if msg != nil {
		log.Println("write:", msg)
		return
	}

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}

}
