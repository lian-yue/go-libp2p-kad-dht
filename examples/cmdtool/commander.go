package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
	"github.com/stefanhans/go-libp2p-kad-dht"
)

var (
	cmdChan = make(chan string)

	commands    = make(map[string]string)
	commandKeys []string

	tmpDebugfile *os.File
)

func commandsInit() {
	commands = make(map[string]string)

	// NODE
	commands["node"] = "node [all] \n\t show information of this node\n" +
		"\t\t - ID\n" +
		"\t\t - Addresses\n" +
		"\t\t - Peerstore\n"

	// BOOTSTRAP
	commands["bootstrap"] = "bootstrap \n\t list bootstrap peers suggested to join\n"
	commands["join"] = "join [<bootstrap> ...] \n\t join bootstrap peer(s)\n" +
		"\n" +
		"\t Connect ensures there is a connection between this host and the peer with\n" +
		"\t given peer.ID. Connect will absorb the addresses in pi into its internal\n" +
		"\t peerstore. If there is not an active connection, Connect will issue a\n" +
		"\t h.Network.Dial, and block until a connection is open, or an error is\n" +
		"\t returned. \n"

	// RENDEZVOUS
	commands["rendezvous"] = "rendezvous \n\t show current rendezvous string with cid set\n"
	commands["setrendezvous"] = "setrendezvous <string> \n\t set cid of string as rendezvous point\n"

	commands["advertise"] = "advertise [<rendezvous> default: as stored by setrendezvous [timeout in seconds, default: 10]] \n\t advertise rendezvous point\n" +
		"\n" +
		"\t Advertise is a utility function that persistently advertises a service through an Advertiser\n" +
		"\t using RoutingDiscovery which is an implementation of discovery using ContentRouting\n"

	// CID
	commands["provide [timeout in seconds, default: 10]"] = "provide announces providing the rendezvous point\n" +
		"\n" +
		"\t Provide makes this node announce that it can provide a value for the given key\n"

	commands["providekey key [timeout in seconds, default: 10]"] = "providekey announces providing the specified key\n" +
		"\n" +
		"\t Provide makes this node announce that it can provide a value for the given key\n"

	commands["kdfindproviders"] = "kdfindproviders \n\t kdfindproviders searches until the context expires\n"
	commands["kdfindprovidersasync"] = "kdfindprovidersasync \n\t kdfindprovidersasync is the same thing as kdfindproviders, but returns a channel. " +
		" \n\t Peers will be returned on the channel as soon as they are found, even before the search query completes.\n"

	// Peer
	commands["kdfindpeer"] = "kdfindpeer [<peerId> default: own node Id [timeout in sec default: 10]]  \n\t kdfindpeer\n"
	commands["kdfindlocal"] = "kdfindlocal \n\t kdfindlocal\n"
	commands["kdupdate"] = "kdupdate \n\t kdupdate signals the routingTable to Update its last-seen status on the given peer.\n"

	// Value
	commands["putvalue"] = "putvalue value [timeout in seconds, default: 10]\n\t putvalue puts a value to the rendezvous point" +
		"\n" +
		"\t PutValue adds value corresponding to given Key.\n" +
		"\t This is the top level \"Store\" operation of the DHT\n"
	commands["putkeyvalue"] = "putkeyvalue key value [timeout in seconds, default: 10]\n\t putvalue puts a key/value pair" +
		"\n" +
		"\t PutValue adds value corresponding to given Key.\n" +
		"\t This is the top level \"Store\" operation of the DHT\n"

	commands["getvalue"] = "getvalue [timeout in seconds, default: 10] \n\t getvalue gets the value of the rendezvous point\n"
	commands["getkeyvalue"] = "getkeyvalue key [timeout in seconds, default: 10] \n\t getkeyvalue gets the value of given key\n"
	commands["getvalues"] = "getvalues \n\t getvalues\n"

	// RoutingTable
	commands["routingtable"] = "routingtable [all] \n\t routingtable shows the routing table with its buckets\n"

	commands["close"] = "close \n\t Close calls Process Close\n"

	// Internals
	commands["debug"] = "debug (on <filename>)|off \n\t debug starts or stops writing debugging output in specified file\n"
	commands["play"] = "play  \n\t for developer playing\n"

	commands["quit"] = "quit  \n\t close the session and exit\n"

	// To store the keys in sorted order
	for commandKey := range commands {
		commandKeys = append(commandKeys, commandKey)
	}
	sort.Strings(commandKeys)
}

// Execute a command specified by the argument string
func executeCommand(commandline string) bool {

	// Trim prefix and split string by white spaces
	commandFields := strings.Fields(commandline)

	// Check for empty string without prefix
	if len(commandFields) > 0 {

		// Switch according to the first word and call appropriate function with the rest as arguments
		switch commandFields[0] {

		case "node":
			nodeInfo(commandFields[1:])
			return true

		case "bootstrap":
			listBootstrap(commandFields[1:])
			return true

		case "join":
			joinBootstrap(commandFields[1:])
			return true

		case "rendezvous":
			rendezvous(commandFields[1:])
			return true

		case "setrendezvous":
			setRendezvous(commandFields[1:])
			return true

		case "advertise":
			advertise(commandFields[1:])
			return true

		// Cid
		case "provide":
			kDhtProvide(commandFields[1:])
			return true

		// Cid
		case "providekey":
			kDhtProvideKey(commandFields[1:])
			return true

		// Cid
		case "kdfindproviders":
			kDhtFindProviders(commandFields[1:])
			return true

		// Cid
		case "kdfindprovidersasync":
			kDhtFindProvidersAsync(commandFields[1:])
			return true

		// Peer
		case "kdfindpeer":
			kDhtFindPeer(commandFields[1:])
			return true

		// Peer
		case "kdfindlocal":
			kDhtFindLocal(commandFields[1:])
			return true

		// Peer
		case "kdupdate":
			kDhtUpdate(commandFields[1:])
			return true

		// Value
		case "putvalue":
			kDhtPutValue(commandFields[1:])
			return true

		case "putkeyvalue":
			kDhtPutKeyValue(commandFields[1:])
			return true

		// Value
		case "getvalue":
			kDhtGetValue(commandFields[1:])
			return true

		case "getkeyvalue":
			kDhtGetKeyValue(commandFields[1:])
			return true

		// Value
		case "getvalues":
			kDhtGetValues(commandFields[1:])
			return true

		// RoutingTable
		case "routingtable":
			kDhtRoutingTable(commandFields[1:])
			return true

		case "close":
			kDhtClose(commandFields[1:])
			return true

		case "debug":
			kDhtDebug(commandFields[1:])
			return true

		case "quit":
			quitCmdTool(commandFields[1:])
			return true

		case "play":
			play(commandFields[1:])
			return true

		default:
			usage()
			return false
		}

		/*    ******** PACKAGE API ********

			func New(ctx context.Context, h host.Host, options ...opts.Option) (*IpfsDHT, error)
			func NewDHT(ctx context.Context, h host.Host, dstore ds.Batching) *IpfsDHT
			func NewDHTClient(ctx context.Context, h host.Host, dstore ds.Batching) *IpfsDHT

			func (dht *IpfsDHT) Bootstrap(ctx context.Context) error
			func (dht *IpfsDHT) Context() context.Context
			func (dht *IpfsDHT) Process() goprocess.Process
			func (dht *IpfsDHT) BootstrapWithConfig(cfg BootstrapConfig) (goprocess.Process, error)
			func (dht *IpfsDHT) BootstrapOnSignal(cfg BootstrapConfig, signal <-chan time.Time) (goprocess.Process, error)



			func (dht *IpfsDHT) FindLocal(id peer.ID) pstore.PeerInfo
			func (dht *IpfsDHT) FindPeer(ctx context.Context, id peer.ID) (_ pstore.PeerInfo, err error)
			func (dht *IpfsDHT) Update(ctx context.Context, p peer.ID)
		!	func (dht *IpfsDHT) GetPublicKey(ctx context.Context, p peer.ID) (ci.PubKey, error)

		!	func (dht *IpfsDHT) FindPeersConnectedToPeer(ctx context.Context, id peer.ID) (<-chan *pstore.PeerInfo, error)



			//func (dht *IpfsDHT) Provide(ctx context.Context, key cid.Cid, brdcst bool) (err error)
			//func (dht *IpfsDHT) FindProviders(ctx context.Context, c cid.Cid) ([]pstore.PeerInfo, error)

			func (dht *IpfsDHT) FindProvidersAsync(ctx context.Context, key cid.Cid, count int) <-chan pstore.PeerInfo



			func (dht *IpfsDHT) PutValue(ctx context.Context, key string, value []byte, opts ...ropts.Option) (err error)
			func (dht *IpfsDHT) GetValue(ctx context.Context, key string, opts ...ropts.Option) (_ []byte, err error)
			func (dht *IpfsDHT) GetValues(ctx context.Context, key string, nvals int) (_ []RecvdVal, err error)

			func (dht *IpfsDHT) SearchValue(ctx context.Context, key string, opts ...ropts.Option) (<-chan []byte, error)
			func (dht *IpfsDHT) GetClosestPeers(ctx context.Context, key string) (<-chan peer.ID, error)



			func (dht *IpfsDHT) Close() error
		*/

	}
	return false
}

// Display the usage of all available commands
func usage() {
	for _, key := range commandKeys {
		fmt.Printf("%v\n", commands[key])
	}

}

func nodeInfo(arguments []string) {

	// Id
	fmt.Printf("%v\t%s\n", node.ID().String(), node.ID().Pretty())

	if len(arguments) > 0 && arguments[0] == "all" {
		for i, addr := range node.Addrs() {

			fmt.Printf("%d: %s\n", i, addr)
		}
		//node.Peerstore().AddAddrs(node.ID(), node.Addrs(), peerstore.AddressTTL)
		pstore := node.Peerstore()

		//fmt.Printf("%v\n", node.Peerstore())
		//fmt.Printf("%v\n", peerstore.Peers())
		fmt.Printf("Peers with addresse(s): %v\n", pstore.PeersWithAddrs())
		fmt.Printf("Peers with key(s): %v\n", pstore.PeersWithKeys())
		//fmt.Printf("%v\n", peerstore.Addrs())

		protocols, err := pstore.GetProtocols(node.ID())
		if err != nil {
			fmt.Printf("error: peerstore.GetProtocols(node.ID()): %v\n", err)
		}
		fmt.Printf("Protocol(s): %v\n", protocols)

	}

}

func listBootstrap(arguments []string) {
	_ = arguments[0]

	for i, p := range bootstrapPeers {
		fmt.Printf("%d: %s\n", i, p)
	}

}

func StringsToAddrs(addrStrings []string) (maddrs []multiaddr.Multiaddr, err error) {
	for _, addrString := range addrStrings {
		addr, err := multiaddr.NewMultiaddr(addrString)
		if err != nil {
			return maddrs, err
		}
		maddrs = append(maddrs, addr)
	}
	return
}

func joinBootstrap(arguments []string) {
	var wg sync.WaitGroup
	pCnt := 0

	if len(arguments) == 0 {

		bootstrapPeerAddrs, err = StringsToAddrs(bootstrapPeers)
		if err != nil {
			panic(err)
		}

		// Let's connect to the bootstrap nodes first. They will tell us about the
		// other nodes in the network.
		for _, peerAddr := range bootstrapPeerAddrs {
			wg.Add(1)
			pCnt++
			peerinfo, _ := peerstore.InfoFromP2pAddr(peerAddr)
			go func() {
				defer wg.Done()
				if err := node.Connect(context.Background(), *peerinfo); err != nil {
					//panic(err)
					fmt.Printf("%s NOT joined\n", *peerinfo)
				} else {
					fmt.Printf("%s joined\n", *peerinfo)
				}
			}()
		}
	} else {

		bootstrapPeerAddrs, err = StringsToAddrs(arguments)
		if err != nil {
			panic(err)
		}
		for _, peerAddr := range bootstrapPeerAddrs {
			wg.Add(1)
			pCnt++
			peerinfo, _ := peerstore.InfoFromP2pAddr(peerAddr)
			go func() {
				defer wg.Done()
				if err := node.Connect(context.Background(), *peerinfo); err != nil {
					panic(err)
				} else {
					fmt.Printf("%s joined\n", *peerinfo)
				}
			}()
		}
	}
	go func() {
		wg.Wait()
		fmt.Printf("Joined with all %d peers\n%s", pCnt, prompt())
	}()
}

func rendezvous(arguments []string) {

	fmt.Printf("%q\n", rendezvousString)

}

func setRendezvous(arguments []string) {

	if len(arguments) > 0 {
		rendezvousPoint, err = v1b.Sum([]byte(arguments[0]))
		if err != nil {
			fmt.Printf("error: could not create rendezvous point: %q\n", arguments[0])
		}
		rendezvousString = arguments[0]
	} else {
		fmt.Printf("no rendezvous string specified: %q\n", strings.Join(arguments, " "))
	}

}

func advertise(arguments []string) {
	ns := rendezvousString
	if len(arguments) > 0 {
		ns = arguments[0]
	}

	if routingDiscovery == nil {
		routingDiscovery = discovery.NewRoutingDiscovery(kademliaDHT)
	}
	discovery.Advertise(ctx, routingDiscovery, ns)

}

func kDhtProvide(arguments []string) {

	timeout := 10
	if len(arguments) > 0 {
		sec, err := strconv.Atoi(arguments[0])
		if err != nil {
			fmt.Printf("%Error: %q is not a valid number: %v\n", arguments[0], err)

			return
		}
		if sec > 0 {
			timeout = sec
		} else {
			fmt.Printf("Error: %d is zero or negative\n", sec)

			return
		}
	}

	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))
	defer cancel()

	err = kademliaDHT.Provide(ctxTimeOut, rendezvousPoint, true)
	if err != nil {
		fmt.Printf("%s %v\n", prompt(), err)
	}
}

func kDhtProvideKey(arguments []string) {

	timeout := 10
	if len(arguments) > 0 {
		sec, err := strconv.Atoi(arguments[0])
		if err != nil {
			fmt.Printf("%Error: %q is not a valid number: %v\n", arguments[0], err)

			return
		}
		if sec > 0 {
			timeout = sec
		} else {
			fmt.Printf("Error: %d is zero or negative\n", sec)

			return
		}
	}

	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))
	defer cancel()

	err = kademliaDHT.Provide(ctxTimeOut, rendezvousPoint, true)
	if err != nil {
		fmt.Printf("%s %v\n", prompt(), err)
	}

}

func kDhtFindProviders(arguments []string) {
	timeout := 10
	if len(arguments) > 0 {
		sec, err := strconv.Atoi(arguments[0])
		if err != nil {
			fmt.Printf("Error: %q is not a valid number: %v\n", arguments[0], err)

			return
		}
		if sec > 0 {
			timeout = sec
		} else {
			fmt.Printf("Error: %d is zero or negative\n", sec)

			return
		}
	}

	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))
	defer cancel()
	peers, err := kademliaDHT.FindProviders(ctxTimeOut, rendezvousPoint)
	if err != nil {
		fmt.Printf("Error: kademliaDHT.FindProviders: %v\n", err)
		return
	}
	for i, p := range peers {
		fmt.Printf("%d\t%v\n", i, p.ID)
		for j, addr := range p.Addrs {
			fmt.Printf("\t%d\t%v\n", j, addr)
		}
	}

}

func kDhtFindProvidersAsync(arguments []string) {
	timeout := 10
	if len(arguments) > 0 {
		sec, err := strconv.Atoi(arguments[0])
		if err != nil {
			fmt.Printf("Error: %q is not a valid number: %v\n", arguments[0], err)

			return
		}
		if sec > 0 {
			timeout = sec
		} else {
			fmt.Printf("Error: %d is zero or negative\n", sec)

			return
		}
	}

	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))
	defer cancel()

	peersChan := kademliaDHT.FindProvidersAsync(ctxTimeOut, rendezvousPoint, 5)

	var p peerstore.PeerInfo
	for l := 0; l < 20; l++ {

		p = <-peersChan
		fmt.Printf("%d: %v\n", l, p.ID)
		for i, addr := range p.Addrs {
			fmt.Printf("\t%d\t%v\n", i, addr)
		}
	}

}

func kDhtFindPeer(arguments []string) {
	timeout := 10
	if len(arguments) > 1 {
		sec, err := strconv.Atoi(arguments[1])
		if err != nil {
			fmt.Printf("%Error: %q is not a valid number: %v\n", arguments[1], err)

			return
		}
		if sec > 0 {
			timeout = sec
		} else {
			fmt.Printf("Error: %d is zero or negative\n", sec)

			return
		}
	}

	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))
	defer cancel()

	peerId := node.ID()

	if len(arguments) > 0 {

		peerId, err = peer.IDFromString(arguments[0])
		if err != nil {
			fmt.Printf("Error: peer.IDFromString: %v\n", err)

			return
		}

	}

	p, err := kademliaDHT.FindPeer(ctxTimeOut, peerId)
	if err != nil {
		fmt.Printf("Error: kademliaDHT.FindPeer: %v\n", err)

		return
	}

	fmt.Printf("%v\t%s\n", p.ID.String(), p.ID.Pretty())
	for i, addr := range p.Addrs {

		fmt.Printf("%d: %s\n", i, addr)
	}

}
func kDhtFindLocal(arguments []string) {
	_ = arguments[0]

	p := kademliaDHT.FindLocal(node.ID())
	fmt.Printf("%v\t%s\n", p.ID.String(), p.ID.Pretty())
	for i, addr := range p.Addrs {

		fmt.Printf("%d: %s\n", i, addr)
	}

}

func kDhtUpdate(arguments []string) {
	timeout := 10
	if len(arguments) > 0 {
		sec, err := strconv.Atoi(arguments[0])
		if err != nil {
			fmt.Printf("Error: %q is not a valid number: %v\n", arguments[0], err)

			return
		}
		if sec > 0 {
			timeout = sec
		} else {
			fmt.Printf("Error: %d is zero or negative\n", sec)

			return
		}
	}

	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))
	defer cancel()
	//peerId, err := peer.IDFromString(arguments[0])
	//if err != nil {
	//	fmt.Printf("Error: peer.IDFromString: %v\n", err)
	//	return
	//}
	kademliaDHT.Update(ctxTimeOut, node.ID())

}

func kDhtPutValue(arguments []string) {
	timeout := 10
	if len(arguments) > 1 {
		sec, err := strconv.Atoi(arguments[1])
		if err != nil {
			fmt.Printf("Error: %q is not a valid number: %v\n", arguments[1], err)

			return
		}
		if sec > 0 {
			timeout = sec
		} else {
			fmt.Printf("Error: %d is zero or negative\n", sec)

			return
		}
	}

	if len(arguments) > 0 {
		ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))
		defer cancel()
		//peerId, err := peer.IDFromString(arguments[0])
		//if err != nil {
		//	fmt.Printf("Error: peer.IDFromString: %v\n", err)
		//	return
		//}
		err = kademliaDHT.PutValue(ctxTimeOut, rendezvousString, []byte(arguments[0]), dht.Quorum(1))
		if err != nil {
			fmt.Printf("Error: kademliaDHT.PutValue: %v\n", err)

			return
		}
	} else {
		fmt.Printf("Error: no value specified\n")

		return
	}

}

func kDhtPutKeyValue(arguments []string) {
	timeout := 10
	if len(arguments) > 2 {
		sec, err := strconv.Atoi(arguments[2])
		if err != nil {
			fmt.Printf("%Error: %q is not a valid number: %v\n", arguments[2], err)

			return
		}
		if sec > 0 {
			timeout = sec
		} else {
			fmt.Printf("Error: %d is zero or negative\n", sec)

			return
		}
	}

	if len(arguments) > 0 {
		ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))
		defer cancel()
		//peerId, err := peer.IDFromString(arguments[0])
		//if err != nil {
		//	fmt.Printf("Error: peer.IDFromString: %v\n", err)
		//	return
		//}
		err = kademliaDHT.PutValue(ctxTimeOut, arguments[0], []byte(arguments[1]), dht.Quorum(1))
		if err != nil {
			fmt.Printf("Error: kademliaDHT.PutValue: %v\n", err)

			return
		}
	} else {
		fmt.Printf("Error: no key and no value specified\n")

		return
	}

}

func kDhtGetValue(arguments []string) {
	timeout := 30
	if len(arguments) > 0 {
		sec, err := strconv.Atoi(arguments[0])
		if err != nil {
			fmt.Printf("Error: %q is not a valid number: %v\n", arguments[0], err)

			return
		}
		if sec > 0 {
			timeout = sec
		} else {
			fmt.Printf("Error: %d is zero or negative\n", sec)

			return
		}
	}

	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))
	defer cancel()
	//peerId, err := peer.IDFromString(arguments[0])
	//if err != nil {
	//	fmt.Printf("Error: peer.IDFromString: %v\n", err)
	//	return
	//}
	val, err := kademliaDHT.GetValue(ctxTimeOut, rendezvousString, dht.Quorum(1))
	if err != nil {
		fmt.Printf("Error: kademliaDHT.GetValue: %v\n", err)

		return
	}
	fmt.Printf("%s\n", val)

}

func kDhtGetKeyValue(arguments []string) {
	timeout := 30
	if len(arguments) > 1 {
		sec, err := strconv.Atoi(arguments[1])
		if err != nil {
			fmt.Printf("Error: %q is not a valid number: %v\n", arguments[1], err)

			return
		}
		if sec > 0 {
			timeout = sec
		} else {
			fmt.Printf("%Error: %d is zero or negative\n", sec)

			return
		}
	}

	if len(arguments) > 0 {
		ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))
		defer cancel()
		//peerId, err := peer.IDFromString(arguments[0])
		//if err != nil {
		//	fmt.Printf("Error: peer.IDFromString: %v\n", err)
		//	return
		//}
		val, err := kademliaDHT.GetValue(ctxTimeOut, arguments[0], dht.Quorum(1))
		if err != nil {
			fmt.Printf("Error: kademliaDHT.GetValue: %v\n", err)

			return
		}
		fmt.Printf("%s\n", val)
	} else {
		fmt.Printf("Error: no key specified\n")

		return
	}

}

func kDhtGetValues(arguments []string) {
	timeout := 30
	if len(arguments) > 0 {
		sec, err := strconv.Atoi(arguments[1])
		if err != nil {
			fmt.Printf("Error: %q is not a valid number: %v\n", arguments[1], err)

			return
		}
		if sec > 0 {
			timeout = sec
		} else {
			fmt.Printf("%Error: %d is zero or negative\n", sec)

			return
		}
	}

	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))
	defer cancel()
	//peerId, err := peer.IDFromString(arguments[0])
	//if err != nil {
	//	fmt.Printf("Error: peer.IDFromString: %v\n", err)
	//	return
	//}
	recvdVals, err := kademliaDHT.GetValues(ctxTimeOut, rendezvousString, 1)
	if err != nil {
		fmt.Printf("Error: kademliaDHT.GetValues: %v\n", err)

		return
	}
	for i, recvdVal := range recvdVals {

		fmt.Printf("%d: %q from %v\n", i, recvdVal.Val, recvdVal.From)
	}

}

func kDhtGetKeyValues(arguments []string) {
	timeout := 30
	if len(arguments) > 1 {
		sec, err := strconv.Atoi(arguments[1])
		if err != nil {
			fmt.Printf("Error: %q is not a valid number: %v\n", arguments[1], err)

			return
		}
		if sec > 0 {
			timeout = sec
		} else {
			fmt.Printf("%Error: %d is zero or negative\n", sec)

			return
		}
	}

	if len(arguments) > 0 {
		ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))
		defer cancel()
		//peerId, err := peer.IDFromString(arguments[0])
		//if err != nil {
		//	fmt.Printf("Error: peer.IDFromString: %v\n", err)
		//	return
		//}
		recvdVals, err := kademliaDHT.GetValues(ctxTimeOut, arguments[0], 1)
		if err != nil {
			fmt.Printf("Error: kademliaDHT.GetValue: %v\n", err)

			return
		}
		for i, recvdVal := range recvdVals {

			fmt.Printf("%d: %q from %v\n", i, recvdVal.Val, recvdVal.From)
		}
	} else {
		fmt.Printf("Error: no key and no value specified\n")

		return
	}

}

func kDhtRoutingTable(arguments []string) {

	routingTable := kademliaDHT.RoutingTable()

	if len(arguments) > 0 && arguments[0] == "all" {

		for i, p := range routingTable.ListPeers() {
			fmt.Printf("%d: %v\n", i, p.String())
		}
	}

	routingTable.Print()

}

func kDhtClose(arguments []string) {
	_ = arguments[0]

	err = kademliaDHT.Close()
	if err != nil {
		panic(err)
	}

}

func kDhtDebug(arguments []string) {

	if len(arguments) == 0 ||
		(len(arguments) == 1 && arguments[0] != "off") {
		fmt.Printf("Error: wrong input. Usage: \n\t 'debug (on <filename>) | off\n")

		return
	}

	if arguments[0] == "on" && len(arguments) > 1 {
		tmpDebugfile, err = startLogging(arguments[1])
		if err != nil {
			fmt.Printf("Error: startLogging: %v\n", err)
		}

		log.Infof("Start debugging ... ")
	}

	if arguments[0] == "off" {
		log.Infof("Stop debugging ")
		_ = tmpDebugfile.Close()
	}

}

func quitCmdTool(arguments []string) {

	// Get rid of warnings
	_ = arguments

	os.Exit(0)
}

func play(arguments []string) {

	fmt.Printf("rendezvousPoint: %q\n", rendezvousPoint)
	fmt.Printf("rendezvousPoint.KeyString(): %q\n", rendezvousPoint.KeyString())

	//rendezvousPoint.Bytes()

	//_ = arguments[0]
	//
	//for i, addr := range node.Addrs() {
	//
	//	fmt.Printf("%d: %s\n", i, addr)
	//	for j, protocol := range addr.Protocols() {
	//
	//		fmt.Printf("\t%d: %s\n", j, protocol.Name)
	//		if protocol.Code == multiaddr.P_IP4 && strings.Contains(addr.String(), "/127.0.0.1/") {
	//			id, err := addr.ValueForProtocol(multiaddr.P_TCP)
	//			if err != nil {
	//				fmt.Printf("Error: addr.ValueForProtocol(multiaddr.P_IP4): %v\n", err)
	//
	//				return
	//			}
	//
	//			myid, err := peer.IDFromString(id)
	//			if err != nil {
	//				fmt.Printf("Error: peer.IDFromString(id): %v\n", err)
	//
	//				return
	//			}
	//			fmt.Printf("myid: %v\n", myid)
	//		}
	//	}
	//}

}
