package structure

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/conformal/btcwire"
	"github.com/jonasnick/btcP2P"
	"gopkg.in/fatih/set.v0"
)

type NetAddressSlice btcP2P.NetAddressSlice

// returns a slice of the known addresses for a peer
// tries to get all known addresses by sending a getaddr
// message multiple times
// the slice is ordered based on the timestamp
// TODO: optimize by sending multiple getaddr requests at once
func KnownAddresses(peer *btcP2P.Peer) chan btcP2P.NetAddressSlice {
	// what is the percentage of known nodes we want to get
	aimPercentageKnown := float32(0.95)

	retChan := make(chan btcP2P.NetAddressSlice)

	go func() {
		addrs := make(map[string]*btcwire.NetAddress)
		for {
			peer.SendSimple(btcwire.NewMsgGetAddr())
			recvAddrs, err := peer.AddrRequestBlocking()
			if err != nil {
				return
			}
			addrsSize := len(addrs)
			for _, k := range recvAddrs {
				addrs[fmt.Sprintf("%s:%d", k.IP, k.Port)] = k
			}
			numNewAddrs := len(addrs) - addrsSize
			percentageKnown := 1 - (float32(numNewAddrs) / float32(len(recvAddrs)))
			log.Println("all addrs", len(addrs), "recv addrs", len(recvAddrs), "new addrs", numNewAddrs, percentageKnown)
			if percentageKnown >= aimPercentageKnown {
				break
			}
		}

		addrsList := make(btcP2P.NetAddressSlice, len(addrs))
		i := 0
		for addr := range addrs {
			addrsList[i] = addrs[addr]
			i += 1
		}
		sort.Sort(addrsList)
		retChan <- addrsList
		close(retChan)
		return
	}()
	return retChan
}

// returns a set of net addresses that have a newer timestamp than t
func selectUntil(s btcP2P.NetAddressSlice, t time.Time) *set.Set {
	ret := set.New()
	for i := len(s) - 1; i >= 0; i-- {
		if s[i].Timestamp.After(t) {
			ret.Add(btcP2P.AddrToString(s[i]))
		} else {
			break
		}
	}
	return ret
}

type KnownAddressStratF func(knownAddresses btcP2P.NetAddressSlice) *set.Set

type KnownAddressStrat struct {
	F    KnownAddressStratF
	Name string
}

// select knownAddresses that are between the newest timestamp and timestamp - timediff
// expects knownAddresses to be sorted
func MakeKnownAddressTimeStrat(timediff time.Duration) *KnownAddressStrat {
	return &KnownAddressStrat{func(knownAddresses btcP2P.NetAddressSlice) *set.Set {
		newestTimestamp := knownAddresses[len(knownAddresses)-1].Timestamp
		return selectUntil(knownAddresses, newestTimestamp.Add(-timediff))

	}, fmt.Sprintf("%d", int(timediff.Minutes()))}
}

// select n knownAddresses
func MakeKnownAddressNumberStrat(n int) *KnownAddressStrat {
	return &KnownAddressStrat{func(knownAddresses btcP2P.NetAddressSlice) *set.Set {
		knownAddressesSlice := knownAddresses[len(knownAddresses)-n : len(knownAddresses)]
		ret := set.New()
		for _, addr := range knownAddressesSlice {
			ret.Add(btcP2P.AddrToString(addr))
		}
		return ret
	}, fmt.Sprintf("%d", n)}
}
