package main

import (
	"os"
	"os/signal"
	"syscall"
	"log"
	"time"
	"fmt"
	"github.com/nats-io/nats.go"
)

const (
	Disconnect = 0
	Reconnect = 1
	Closed = 2
	Discovered = 3
)

func main() {
	// Signal Channel
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Notification channel
	done := make(chan bool)

	nc, err := nats.Connect(nats.DefaultURL,
    		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
        		log.Printf("client disconnected: %v", err)
    		}),
    		nats.ReconnectHandler(func(_ *nats.Conn) {
        		log.Printf("client reconnected")
    		}),
    		nats.ClosedHandler(func(_ *nats.Conn) {
        		log.Printf("client closed")
    		}),
    		nats.DiscoveredServersHandler(func(nc *nats.Conn) {
        		log.Printf("Known servers: %v\n", nc.Servers())
        		log.Printf("Discovered servers: %v\n", nc.DiscoveredServers())
    		}),
	)

	if err != nil {
    		log.Fatal(err)
	}

	// create new Tickers
	tick := time.NewTicker(10 * time.Millisecond)
	boom := time.NewTicker(50 * time.Millisecond)

	// Create timer to end loop
	end  := time.After(10000 * time.Millisecond)

	go func() {
		for {
			select {
			case <-tick.C:
				fmt.Println("tick.")
				if err := nc.Publish("foo", []byte("All is well ticky-tock!")); err != nil {
					log.Fatal(err)
				}
			case <-boom.C:
				fmt.Println("BOOM!")
				if err := nc.Publish("foo", []byte("BOOM BOOM BOOM!")); err != nil {
					log.Fatal(err)
				}
			case <-end:
			case <-sigs:
				fmt.Println("Shutting down.")
				tick.Stop()
				boom.Stop()

				time.Sleep(2 * time.Millisecond)

				nc.Flush()
				nc.Drain()
				done <- true
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()

	<-done

	log.Println("Done.")

	nc.Close()
	os.Exit(0)
}
