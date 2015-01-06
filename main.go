package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/conformal/btcrpcclient"
	"github.com/jonasnick/btcP2P"
	"github.com/jonasnick/btcP2PStruct/structure"
	"gopkg.in/fatih/set.v0"
)

// queries a bitcoin node for known peer addresses via p2p
// and via rpc and computes the intersection

/*
 * adjust these settings
 */
var ownIpAddress = "127.0.0.1:8333"

// p2p
var nodeAddress = "127.0.0.1:8333"

// rpc
var connCfg = &btcrpcclient.ConnConfig{
	Host:         "localhost:8332",
	User:         "user",
	Pass:         "1234",
	HttpPostMode: true, // Bitcoin core only supports HTTP POST mode
	DisableTLS:   true, // Bitcoin core does not provide TLS by default
}

func mean(s []float32) float32 {
	sum := float32(0)
	for _, ele := range s {
		sum += ele
	}
	return sum / float32(len(s))
}

type Stat struct {
	Tp int
	Fp int
	Fn int
	// numRpcPeers = tp + fn
	NumKnownPeers int
}

func (s *Stat) precision() float64 {
	return float64(s.Tp) / float64(s.Tp+s.Fp)
}

func (s *Stat) recall() float64 {
	return float64(s.Tp) / float64(s.Tp+s.Fn)
}

func knownAddressStatistics(strats []*structure.KnownAddressStrat, knownAddresses btcP2P.NetAddressSlice, rpcPeers *set.Set) []*Stat {
	ret := make([]*Stat, len(strats))
	for i, strat := range strats {
		selected := strat.F(knownAddresses)
		tp := set.Intersection(selected, rpcPeers).Size()
		fp := selected.Size() - tp
		// the local connected node is not included in the known nodes, ergo not in selected
		fn := rpcPeers.Size() - 1 - tp
		ret[i] = &Stat{tp, fp, fn, len(knownAddresses)}
	}
	return ret
}

func p2pConnect(stateResultChan chan *btcP2P.PeerState) *btcP2P.Peer {
	var peer *btcP2P.Peer

	for {
		newConnectionChan := make(chan *btcP2P.Connection)

		// connect to node via p2p bitcoin protocol
		go btcP2P.Connect(nodeAddress, newConnectionChan)
		p2pConn := <-newConnectionChan
		if p2pConn.Err != nil {
			log.Println(p2pConn.Err)
			<-time.Tick(5 * time.Second)
			continue
		}
		peer = btcP2P.NewPeer(p2pConn, stateResultChan)
		go peer.BasicHandler()
		go peer.NegotiateVersionHandler(ownIpAddress)
		<-stateResultChan
		break
	}

	return peer
}

func logSomeAddresses(knownAddresses btcP2P.NetAddressSlice, rpcPeers *set.Set) {
	some := 25
	str := ""
	str += "known addresses: "
	for i := len(knownAddresses) - some - 1; i < len(knownAddresses); i++ {
		addr := knownAddresses[i]
		str += fmt.Sprintf("%s:%d (%v),", addr.IP, addr.Port, addr.Timestamp)
	}
	log.Println(str)
	str = "rpc peers: "
	for _, addr := range set.StringSlice(rpcPeers) {
		str += fmt.Sprintf("%s, ", addr)
	}
	log.Println(str)
}

func writeStats(strats []*structure.KnownAddressStrat, knownAddresses btcP2P.NetAddressSlice, rpcPeers *set.Set, suffix string) {
	stats := knownAddressStatistics(strats, knownAddresses, rpcPeers)

	t := time.Now()
	filename := fmt.Sprintf("data/%d-%02d-%02d_%02d:%02d:%02d_%s.csv",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), suffix)

	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString("threshold, tp, fp, fn, numKnownPeers\n")
	if err != nil {
		log.Fatal(err)
	}

	for i, s := range stats {
		_, err := f.WriteString(fmt.Sprintf("%s, %d, %d, %d, %d, %d\n", strats[i].Name, s.Tp, s.Fp, s.Fn, s.NumKnownPeers))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func handleKnownAddresses(rpcClient *btcrpcclient.Client, knownAddresses btcP2P.NetAddressSlice) {
	peerInfos, err := rpcClient.GetPeerInfo()
	if err != nil {
		log.Fatal(err)
	}
	rpcPeers := set.New()
	for _, peerInfo := range peerInfos {
		rpcPeers.Add(peerInfo.Addr)
	}

	numberStrats := make([]*structure.KnownAddressStrat, 0)
	for i := 1; i < 10; i++ {
		numberStrats = append(numberStrats, structure.MakeKnownAddressNumberStrat(i))
	}
	for i := 10; i < 100; i += 5 {
		numberStrats = append(numberStrats, structure.MakeKnownAddressNumberStrat(i))
	}

	timeStrats := make([]*structure.KnownAddressStrat, 0)
	for i := 10; i < 200; i += 10 {
		timeStrats = append(timeStrats, structure.MakeKnownAddressTimeStrat(time.Duration(i)*time.Minute))
	}
	for i := 200; i < 500; i += 40 {
		timeStrats = append(timeStrats, structure.MakeKnownAddressTimeStrat(time.Duration(i)*time.Minute))
	}

	writeStats(numberStrats, knownAddresses, rpcPeers, "number")
	writeStats(timeStrats, knownAddresses, rpcPeers, "time")

	log.Printf("completed round")
	logSomeAddresses(knownAddresses, rpcPeers)
	<-time.After(60 * time.Minute)
}

func main() {
	f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	log.SetOutput(f)

	// connect to node via rpc
	rpcClient, err := btcrpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer rpcClient.Shutdown()

	stateResultChan := make(chan *btcP2P.PeerState)

	var peer *btcP2P.Peer
	peer = p2pConnect(stateResultChan)

	for {
		select {
		case knownAddresses := <-structure.KnownAddresses(peer):
			handleKnownAddresses(rpcClient, knownAddresses)
		case peerState := <-stateResultChan:
			_, state := peerState.Get()
			switch state.(type) {
			case *btcP2P.StateDisconnected:
				peer = p2pConnect(stateResultChan)
			}
		}
	}
}
