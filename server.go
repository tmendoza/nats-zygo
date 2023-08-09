package main

import (
	"log"
//	"time"
//	"fmt"
	"github.com/nats-io/nats.go"
)

const (
	Disconnect = 0
	Reconnect = 1
	Closed = 2
	Discovered = 3
)

func main() {
	// Notification channel
	notify := make(chan int)

	nc, err := nats.Connect(nats.DefaultURL,
    		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
        		log.Printf("client disconnected: %v", err)
			notify <- Disconnect
    		}),
    		nats.ReconnectHandler(func(_ *nats.Conn) {
        		log.Printf("client reconnected")
			notify <- Reconnect
    		}),
    		nats.ClosedHandler(func(_ *nats.Conn) {
        		log.Printf("client closed")
			notify <- Closed
    		}),
    		nats.DiscoveredServersHandler(func(nc *nats.Conn) {
        		log.Printf("Known servers: %v\n", nc.Servers())
        		log.Printf("Discovered servers: %v\n", nc.DiscoveredServers())
			notify <- Discovered
    		}),
	)

	if err != nil {
    		log.Fatal(err)
	}

	defer nc.Close()

	// Channel Subscriber
	ch := make(chan *nats.Msg, 128)
	sub, err := nc.ChanSubscribe("foo", ch)

	sub.SetPendingLimits(-1, -1)

	// handle err
	if err != nil {
		log.Fatalf("error occured: %v\n", err)
	}

	//tick := time.Tick(1000 * time.Millisecond)
	//boom := time.After(5000 * time.Millisecond)

	for {
		select {
		case m := <-ch:
			log.Printf("Got a message: %s\n", string(m.Data))
		//case <-tick:
		//	fmt.Println("tick.")
		//case <-boom:
		//	fmt.Println("BOOM!")
		case n := <-notify:
			switch n {
				case Reconnect:
					log.Println("reconnect event retrieved")
				case Disconnect:
					log.Println("disconnect event retrieved")
				case Closed:
					log.Println("closed event retrieved")
				case Discovered:
					log.Println("discovery event retrieved")
			}
		default:
			//time.Sleep(50 * time.Millisecond)
		}
	}

	sub.Unsubscribe()
	nc.Drain()

	close(ch)
}
