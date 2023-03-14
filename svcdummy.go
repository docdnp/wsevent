package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/textileio/go-threads/broadcast"
)

var b broadcast.Broadcaster

func produceDummyEvents(addr string) {

	rand.Seed(time.Now().UnixNano())

	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	RandStringRunes := func(n int) string {
		b := make([]rune, n)
		for i := range b {
			b[i] = letterRunes[rand.Intn(len(letterRunes))]
		}
		return string(b)
	}

	i := 0
	for {
		time.Sleep(100 * time.Millisecond)
		b.Send(strconv.Itoa(i) + ": " + addr + ": " + RandStringRunes(10) + "\n")
		i += 1
	}
}

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func serveClient(w http.ResponseWriter, r *http.Request) {
	l := b.Listen()
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	defer l.Discard()
	for {
		msg := <-l.Channel()
		msgPayload, ok := msg.(string)
		if !ok {
			log.Print("warning: skipping unknown data")
			continue
		}
		err = c.WriteMessage(websocket.TextMessage, []byte(msgPayload))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	flag.Parse()
	fmt.Println("Hello, World!")
	go produceDummyEvents(*addr)
	http.HandleFunc("/echo", serveClient)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
