package main

import (
	"log"
	"time"
	"github.com/nats-io/nats.go"
)

func main() {
	
	log.Println("onnecting to Default NATS URL...")

	nc, err := nats.Connect(nats.DefaultURL)

	if err != nil {
		log.Fatal("Could not connect to Default URL")
	}

	defer nc.Close()

	if nc.Status() != nats.CONNECTED || nc.Status().String() != "CONNECTED" {
		log.Fatal("Should have status set to CONNECTED")
	}

	msg := []byte("This is a test string specialty.")

	for i := 0; i < 10000; i++ {
		if err := nc.Publish("foo", msg); err != nil {
			log.Fatalf("Error during Publish: %v\n", err)
		}
		time.Sleep(300)
	}

	nc.Flush()
	nc.Drain()
}
