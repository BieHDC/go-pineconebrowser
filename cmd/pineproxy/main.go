package main

import (
	"encoding/hex"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	pineconeMulticast "github.com/matrix-org/pinecone/multicast"
	pineconeRouter "github.com/matrix-org/pinecone/router"

	pb "biehdc.pineconebrowser"
)

var (
	username  = flag.String("name", "Generic Pineman", "your name on the network")
	webui     = flag.String("webui", "127.0.0.1:4200", "address of the built in web ui (set to NONE to turn off for headless)")
	proxyaddr = flag.String("proxy", "127.0.0.1:8088", "where the proxy should listen at (set your http_proxy to this and disable https-only mode in your browser)")
	keyfile   = flag.String("key", "privatekey.keyfile", "the host identity key (will be created if not exists)")
)

func main() {
	flag.Parse()

	//logger := log.New(os.Stdout, "", log.Ltime|log.Lshortfile) //too much cluttering //debug
	logger := log.New(os.Stdout, "", log.Ltime)

	pk, sk, err := pb.GetKeysFromKeyfile(*keyfile)
	if err != nil {
		panic(err)
	}
	pineconeaddress := hex.EncodeToString(pk)
	logger.Println("Pinecone Browser Proxy is ready...", pineconeaddress)

	pRouter := pineconeRouter.NewRouter(nil, sk) // only logger for debug: nil->logger
	pRouter.EnableHopLimiting()                  // do we need or want this?
	//pRouter.EnableWakeupBroadcasts()             // do we need or want this?

	pMulticast := pineconeMulticast.NewMulticast(logger, pRouter)
	pMulticast.Start() //disabled by default due to reasons. needs to be evaluated!

	// Pinehost enabled instances
	pinehost := pb.PinehostSpawn(logger, pRouter, pineconeaddress, *username, "NONE")

	// Now make a webui with all known peers, if they are a pinehost, with clickable elements and etc
	proxyaddr := pinehost.SpawnForwardingProxy(*proxyaddr)
	if proxyaddr != nil {
		logger.Printf("Proxy open on http://%s\n", proxyaddr.String())
	} else {
		logger.Println("Proxy has not been started")
	}

	if *webui != "NONE" {
		webuiaddr := pinehost.SpawnWebUI(*webui)
		if webuiaddr != nil {
			logger.Printf("GUI open on http://%s\n", webuiaddr.String())
		} else {
			logger.Println("GUI has not been started")
		}
	}

	//go watchEvents(pRouter, logger) // bandwidth report

	logger.Println("waiting for ctrl+c")
	waitForExit()
	logger.Println("exiting")
}

// Make sure we can exit cleanly
// Blocks until a signal enters
func waitForExit() {
	closechannel := make(chan os.Signal, 1)
	signal.Notify(closechannel,
		os.Interrupt,
		os.Kill,
		syscall.SIGABRT,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	<-closechannel
}
