package pineweb

import (
	pineconeRouter "github.com/matrix-org/pinecone/router"
	pineconeEvents "github.com/matrix-org/pinecone/router/events"
	"log"
	"strconv"
)

func watchEvents(router *pineconeRouter.Router, logger *log.Logger) {
	eventchannel := make(chan pineconeEvents.Event, 1)
	router.Subscribe(eventchannel)
	for event := range eventchannel {
		switch e := event.(type) {
		// Below are ignored
		case pineconeEvents.PeerAdded:
		case pineconeEvents.PeerRemoved:
		case pineconeEvents.TreeParentUpdate:
		case pineconeEvents.SnakeDescUpdate:
		case pineconeEvents.TreeRootAnnUpdate:
		case pineconeEvents.SnakeEntryAdded:
		case pineconeEvents.SnakeEntryRemoved:
		case pineconeEvents.BroadcastReceived:
		// Above are ignored
		case pineconeEvents.BandwidthReport:
			logger.Println("Bandwith Report:")
			for name, peer := range e.Peers {
				logger.Println("\tPeer: ", name)
				logger.Println("\t\tProtocol: Rx: ", bytestohuman(peer.Protocol.Rx), " - Tx: ", bytestohuman(peer.Protocol.Tx))
				logger.Println("\t\tOverlay: Rx: ", bytestohuman(peer.Overlay.Rx), " - Tx: ", bytestohuman(peer.Overlay.Tx))
			}
			logger.Println()
		default:
			logger.Print("unknown pinecone event")
		}
	}
}

func PineconeEventToString(e pineconeEvents.Event) string {
	switch e.(type) {
	case pineconeEvents.PeerAdded:
		return "PeerAdded"
	case pineconeEvents.PeerRemoved:
		return "PeerRemoved"
	case pineconeEvents.TreeParentUpdate:
		return "TreeParentUpdate"
	case pineconeEvents.SnakeDescUpdate:
		return "SnakeDescUpdate"
	case pineconeEvents.TreeRootAnnUpdate:
		return "TreeRootAnnUpdate"
	case pineconeEvents.SnakeEntryAdded:
		return "SnakeEntryAdded"
	case pineconeEvents.SnakeEntryRemoved:
		return "SnakeEntryRemoved"
	case pineconeEvents.BroadcastReceived:
		return "BroadcastReceived"
	case pineconeEvents.BandwidthReport:
		return "BandwidthReport"
	default:
		return "Unknown Pinecone Event"
	}
}

func bytestohuman(size uint64) string {
	if size < 1024 {
		return strconv.FormatUint(size, 10) + " bytes"
	}
	size = size / 1024
	if size < 1024 {
		return strconv.FormatUint(size, 10) + " kbytes"
	}
	size = size / 1024
	if size < 1024 {
		return strconv.FormatUint(size, 10) + " mbytes"
	}
	size = size / 1024
	if size < 1024 {
		return strconv.FormatUint(size, 10) + " gbytes"
	}
	size = size / 1024
	if size < 1024 {
		return strconv.FormatUint(size, 10) + " tbytes"
	}
	return strconv.FormatUint(size, 10) + " idk what you are doing"
}
