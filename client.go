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

var addr = flag.String("addr", "192.168.0.17:8080", "http service address")

func main() {
	//Main function of client app.
	//This function is called when app start

	file_to_send, err := ioutil.ReadFile("data.go")
	//Reading source code from code file

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	//Setting up notification of connection interrupt

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	//Open websocket connectin with error handling
	defer c.Close()
	//Make sure the connectin will be closed before application exit

	done := make(chan struct{})

	awaitingResults := false
	awaitingRaport := false

	mes := ""

	go func() {
		//Incoming mesasges handler - parallel running function
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			//Incomming messages reader with error handling
			mes = string(message)
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
					//Incoming program output handler
					log.Println("Received results:\n", mes)
					awaitingRaport = true
				} else {
					//incoming report handle
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
	//Sending source code to server with error handling

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			//Connection interrupted handler
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
