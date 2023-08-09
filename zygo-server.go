package main

import (
	"flag"
	"log"
//	"time"
	"fmt"
	"os"

	"github.com/nats-io/nats.go"
	"github.com/glycerine/zygomys/v6/zygo"
)

const (
	Disconnect = 0
	Reconnect = 1
	Closed = 2
	Discovered = 3
)

func usage(myflags *flag.FlagSet) {
	log.Printf("zygo command line help:\n")
	myflags.PrintDefaults()
	os.Exit(1)
}

func main() {
	cfg := zygo.NewZlispConfig("zygo")
	cfg.DefineFlags()

	err := cfg.Flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		usage(cfg.Flags)
	}

	if err != nil {
		panic(err)
	}

	err = cfg.ValidateConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "zygo command line error: '%v'\n", err)
		usage(cfg.Flags)
	}

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
	sub, err := nc.ChanQueueSubscribe("foo", "worker", ch)

	sub.SetPendingLimits(-1, -1)

	// handle err
	if err != nil {
		log.Fatalf("error occured: %v\n", err)
	}

	go func() {
		for {
			select {
			case m := <-ch:
				log.Printf("Got a message: %s\n", string(m.Data))
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
			}
		}
	}()

	// the library does all the heavy lifting.
	//zygo.ReplMain(cfg)

	var env *zygo.Zlisp

	if cfg.LoadDemoStructs {
		zygo.RegisterDemoStructs()
	}
	if cfg.Sandboxed {
		env = zygo.NewZlispSandbox()
	} else {
		env = zygo.NewZlisp()
	}
	env.StandardSetup()
	if cfg.LoadDemoStructs {
		// avoid data conflicts by only loading these in demo mode.
		env.ImportDemoData()
	}

	env.AddFunction("slowb", GetText)

	zygo.Repl(env, cfg)

	sub.Unsubscribe()
	nc.Drain()

	close(ch)
}

func GetText(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
    	// ...
	return &zygo.SexpStr{S: "Hello World"}, nil
}
