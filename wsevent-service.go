package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/textileio/go-threads/broadcast"
)

// EventGenerator creates random strings and broadcasts them to any
// connected channel (via broadcast.Broadcaster)
func EventGenerator(bc *broadcast.Broadcaster, serve_address string) {
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
		message := strconv.Itoa(i) + ": " + serve_address + ": " + RandStringRunes(10) + "\n"
		log.Print("Created dummy message: " + message)
		bc.Send(message)
		i += 1
	}
}

// EventConsumer retries connecting to a given event source, processes
// incoming events and broadcasts the to any connected channel
// (via broadcast.Broadcaster)
func EventConsumer(bc *broadcast.Broadcaster, address string, path string) {
	is_running := true
	for {
		retry := make(chan struct{})
		exitnow := make(chan struct{})
		go func() {
			log.Println("Reconnect.. Running? " + strconv.FormatBool(is_running))
			if !is_running {
				close(exitnow)
			}
			defer func() { close(retry) }()

			done := make(chan struct{})
			interrupt := make(chan os.Signal, 1)
			cancelled := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)

			go func() {
				sig := <-interrupt
				is_running = false
				cancelled <- sig
			}()

			u := url.URL{Scheme: "ws", Host: address, Path: path}
			log.Printf("connecting to %s", u.String())

			c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

			if err != nil {
				log.Println("dial:", err)
				return
			}
			defer c.Close()

			go func(c *websocket.Conn) {
				defer close(done)
				for is_running {
					_, message, err := c.ReadMessage()
					if err != nil {
						log.Println("read:", err)
						return
					}
					bc.Send(string(message))
					log.Printf("recv: (%v) %s", is_running, message)
				}
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
					return
				}
			}(c)

			for {
				select {
				case <-done:
					return
				case <-cancelled:
					log.Println("interrupt")
					select {
					case <-done:
					case <-time.After(time.Second):
					}
					close(exitnow)
					return
				}
			}
		}()
		select {
		case <-retry:
			log.Println("Retrying to reconnect to " + address + "/" + path + ".")
			time.Sleep(time.Second)
		case <-exitnow:
			log.Println("Stopping service.")
			os.Exit(0)
		}

	}
}

// EventBroadcaster serves multiple websocket connections and passes incoming
// events (received through broadcast.Broadcaster) to every connected client.
func EventBroadcaster(b *broadcast.Broadcaster) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := b.Listen()

		var upgrader = websocket.Upgrader{} // use default options
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		c, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()
		defer l.Discard()

		hostname, err := os.Hostname()
		if err != nil {
			fmt.Println(err)
			return
		}

		for {
			msg := <-l.Channel()
			msgPayload, ok := msg.(string)
			if !ok {
				log.Print("warning: skipping unknown data")
				continue
			}
			err = c.WriteMessage(websocket.TextMessage, []byte(hostname+"<- "+msgPayload))
			if err != nil {
				log.Println("write: ", err)
				break
			}
		}
	}
}

// CliArgs are the application's CLI arguments
type CliArgs struct {
	serve_endpoint   *string
	serve_addr       *string
	serve_path       *string
	consume_endpoint *string
	consume_addr     *string
	consume_path     *string
	no_prod          *bool
}

// ArgumentParser returns the application's configuration.
// First environment variables are used. If CLI args exist
// they superseed the environment
func ArgumentParser() CliArgs {
	fromEnv := func(envvar_name string, default_val interface{}) interface{} {
		val, ok := os.LookupEnv(envvar_name)
		if ok {
			switch default_val.(type) {
			case bool:
				return os.Getenv(envvar_name) != ""
			case string:
				return val
			default:
				return ""
			}
		} else {
			return default_val
		}
	}

	flags := CliArgs{
		serve_endpoint:   flag.String("serve", fromEnv("SERVE_ENDPOINT", "").(string), "http service endpoint (address plus path)"),
		serve_addr:       flag.String("serve-address", fromEnv("SERVE_ADDRESS", "localhost:8080").(string), "http service serve address"),
		serve_path:       flag.String("serve-path", fromEnv("SERVE_PATH", "events").(string), "http servce path"),
		consume_endpoint: flag.String("consume", fromEnv("CONSUME_ENDPOINT", "").(string), "http service endpoint (address plus path)"),
		consume_addr:     flag.String("consume-address", fromEnv("CONSUME_ADDRESS", "").(string), "http service serve address"),
		consume_path:     flag.String("consume-path", fromEnv("CONSUME_PATH", "").(string), "http consume path"),
		no_prod:          flag.Bool("no-produce", fromEnv("NO_PRODUCE", false).(bool), "Deactivate producing. Activate consuming"),
	}

	caps2args := func(caps [][]string, args ...*string) {
		if len(caps) == 0 || len(caps[0]) != len(args)+1 {
			return
		}
		for i, ptr := range args {
			*ptr = caps[0][i+1]
		}
	}

	flag.Parse()

	r := regexp.MustCompile(`^(.*?)/(.*)?$`)
	caps2args(r.FindAllStringSubmatch(*flags.serve_endpoint, -1), flags.serve_addr, flags.serve_path)
	caps2args(r.FindAllStringSubmatch(*flags.consume_endpoint, -1), flags.consume_addr, flags.consume_path)

	return flags
}

func main() {
	var b broadcast.Broadcaster
	flags := ArgumentParser()

	if !*flags.no_prod {
		fmt.Println("Starting as producer...")
		go EventGenerator(&b, *flags.serve_addr)
	} else {
		fmt.Println("Starting as proxy service...")
		fmt.Println("Consuming from: " + *flags.consume_addr + "/" + *flags.serve_path)
		go EventConsumer(&b, *flags.consume_addr, "/"+*flags.consume_path)
	}

	fmt.Println("Serving on: " + *flags.serve_addr + "/" + *flags.serve_path)
	http.HandleFunc("/"+*flags.serve_path, EventBroadcaster(&b))
	log.Fatal(http.ListenAndServe(*flags.serve_addr, nil))
}
