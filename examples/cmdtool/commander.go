package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
)

var (
	writeChan = make(chan string)

	commands    = make(map[string]string)
	commandKeys []string
)

func commandsInit() {
	commands = make(map[string]string)
	commands["self"] = "self \n\t show information of this node\n"
	commands["bootstrap"] = "bootstrap \n\t list bootstrap peer(s)\n"
	commands["join"] = "join [<bootstrap> ...] \n\t join bootstrap peer(s)\n"
	commands["rendezvous"] = "rendezvous \n\t show last rendezvous string\n"
	commands["setrendezvous"] = "setrendezvous <string> \n\t set cid of string as rendezvous point\n"
	commands["advertise"] = "advertise \n\t advertise rendezvous point\n"

	// Cid
	commands["kdprovide"] = "kdprovide \n\t kdprovide makes this node announce that it can provide a value for the given key\n"
	commands["kdfindproviders"] = "kdfindproviders \n\t kdfindproviders searches until the context expires\n"
	commands["kdfindprovidersasync"] = "kdfindprovidersasync \n\t kdfindprovidersasync is the same thing as kdfindproviders, but returns a channel. " +
		" \n\t Peers will be returned on the channel as soon as they are found, even before the search query completes.\n"

	// Peer
	commands["kdfindpeer"] = "kdfindpeer \n\t kdfindpeer\n"
	commands["kdfindlocal"] = "kdfindlocal \n\t kdfindlocal\n"
	commands["kdupdate"] = "kdupdate \n\t kdupdate signals the routingTable to Update its last-seen status on the given peer.\n"

	// Value
	commands["kdputvalue"] = "kdputvalue \n\t kdputvalue\n"
	commands["kdgetvalue"] = "kdgetvalue \n\t kdgetvalue\n"
	commands["kdgetvalues"] = "kdgetvalues \n\t kdgetvalues\n"

	commands["kdclose"] = "kdclose \n\t kdclose\n"

	//commands["debug"] = "\\debug on|off"
	//commands["dht"] = "\\dht"
	commands["quit"] = "quit  \n\t close the session and exit\n"

	// To store the keys in sorted order
	for commandKey := range commands {
		commandKeys = append(commandKeys, commandKey)
	}
	sort.Strings(commandKeys)
}

// Execute a command specified by the argument string
func executeCommand(commandline string) {

	// Trim prefix and split string by white spaces
	commandFields := strings.Fields(commandline)

	// Check for empty string without prefix
	if len(commandFields) > 0 {
		//log.Printf("Command: %q\n", commandFields[0])
		//log.Printf("Arguments (%v): %v\n", len(commandFields[1:]), commandFields[1:])

		// Switch according to the first word and call appropriate function with the rest as arguments
		switch commandFields[0] {

		case "self":
			self(commandFields[1:])

		case "bootstrap":
			listBootstrap(commandFields[1:])

		case "join":
			joinBootstrap(commandFields[1:])

		case "rendezvous":
			rendezvous(commandFields[1:])

		case "setrendezvous":
			setRendezvous(commandFields[1:])

		case "advertise":
			advertise(commandFields[1:])

		// Cid
		case "kdprovide":
			kDhtProvide(commandFields[1:])

		// Cid
		case "kdfindproviders":
			kDhtFindProviders(commandFields[1:])

		// Cid
		case "kdfindprovidersasync":
			kDhtFindProvidersAsync(commandFields[1:])

		// Peer
		case "kdfindpeer":
			kDhtFindPeer(commandFields[1:])

		// Peer
		case "kdfindlocal":
			kDhtFindLocal(commandFields[1:])

		// Peer
		case "kdupdate":
			kDhtUpdate(commandFields[1:])

		// Value
		case "kdputvalue":
			kDhtPutValue(commandFields[1:])

		// Value
		case "kdgetvalue":
			kDhtGetValue(commandFields[1:])

		// Value
		case "kdgetvalues":
			kDhtGetValues(commandFields[1:])

		case "kdclose":
			kDhtClose(commandFields[1:])

		case "quit":
			quitChat(commandFields[1:])

		default:
			usage()
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

	} else {
		fmt.Printf(prompt)
	}
}

// Display the usage of all available commands
func usage() {
	// enhance: order not deterministic bug
	for _, key := range commandKeys {
		fmt.Printf("%v\n", commands[key])
	}
	fmt.Printf(prompt)
}

func self(arguments []string) {

	fmt.Printf("%v\t%s\n", node.ID().String(), node.ID().Pretty())
	for i, addr := range node.Addrs() {

		fmt.Printf("%d: %s\n", i, addr)
	}
	//node.Peerstore().AddAddrs(node.ID(), node.Addrs(), peerstore.AddressTTL)
	fmt.Printf("%v\n", node.Peerstore())

	fmt.Printf(prompt)
}

func listBootstrap(arguments []string) {

	for i, peer := range bootstrapPeers {
		fmt.Printf("%d: %s\n", i, peer)
	}
	fmt.Printf(prompt)

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
					panic(err)
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
		fmt.Printf("Joined with all %d peers\n%s", pCnt, prompt)
	}()
}

func rendezvous(arguments []string) {

	fmt.Printf("%q\n", rendezvousString)
	fmt.Printf(prompt)
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
	fmt.Printf(prompt)
}

func advertise(arguments []string) {

	if routingDiscovery == nil {
		routingDiscovery = discovery.NewRoutingDiscovery(kademliaDHT)
	}
	discovery.Advertise(ctx, routingDiscovery, rendezvousString)

	fmt.Printf(prompt)
}

func kDhtProvide(arguments []string) {
	// todo parameterize timeout
	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(10))
	defer cancel()
	err = kademliaDHT.Provide(ctxTimeOut, rendezvousPoint, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf(prompt)
}

func kDhtFindProviders(arguments []string) {
	// todo parameterize timeout
	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(10))
	defer cancel()
	peers, err := kademliaDHT.FindProviders(ctxTimeOut, rendezvousPoint)
	if err != nil {
		fmt.Printf("Error: kademliaDHT.FindProviders: %v\n", err)
		return
	}
	for i, peer := range peers {
		fmt.Printf("%d\t%v\n", i, peer.ID)
		for j, addr := range peer.Addrs {
			fmt.Printf("\t%d\t%v\n", j, addr)
		}
	}
	fmt.Printf(prompt)
}

func kDhtFindProvidersAsync(arguments []string) {
	// todo parameterize timeout
	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(10))
	defer cancel()

	peersChan := kademliaDHT.FindProvidersAsync(ctxTimeOut, rendezvousPoint, 5)

	var peer peerstore.PeerInfo
	for l := 0; l < 20; l++ {

		peer = <-peersChan
		fmt.Printf("%d: %v\n", l, peer.ID)
		for i, addr := range peer.Addrs {
			fmt.Printf("\t%d\t%v\n", i, addr)
		}
	}
	fmt.Printf(prompt)
}

func kDhtFindPeer(arguments []string) {
	// todo parameterize timeout
	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(10))
	defer cancel()
	//peerId, err := peer.IDFromString(arguments[0])
	//if err != nil {
	//	fmt.Printf("Error: peer.IDFromString: %v\n", err)
	//	fmt.Printf(prompt)
	//	return
	//}
	peer, err := kademliaDHT.FindPeer(ctxTimeOut, node.ID())
	if err != nil {
		fmt.Printf("Error: kademliaDHT.FindPeer: %v\n", err)
		fmt.Printf(prompt)
		return
	}

	fmt.Printf("%v\t%s\n", peer.ID.String(), peer.ID.Pretty())
	for i, addr := range peer.Addrs {

		fmt.Printf("%d: %s\n", i, addr)
	}
	fmt.Printf(prompt)
}
func kDhtFindLocal(arguments []string) {
	peer := kademliaDHT.FindLocal(node.ID())
	fmt.Printf("%v\t%s\n", peer.ID.String(), peer.ID.Pretty())
	for i, addr := range peer.Addrs {

		fmt.Printf("%d: %s\n", i, addr)
	}
	fmt.Printf(prompt)
}

func kDhtUpdate(arguments []string) {
	// todo parameterize timeout
	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(10))
	defer cancel()
	//peerId, err := peer.IDFromString(arguments[0])
	//if err != nil {
	//	fmt.Printf("Error: peer.IDFromString: %v\n", err)
	//	return
	//}
	kademliaDHT.Update(ctxTimeOut, node.ID())
	fmt.Printf(prompt)
}

func kDhtPutValue(arguments []string) {
	// todo parameterize timeout
	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(10))
	defer cancel()
	//peerId, err := peer.IDFromString(arguments[0])
	//if err != nil {
	//	fmt.Printf("Error: peer.IDFromString: %v\n", err)
	//	return
	//}
	err = kademliaDHT.PutValue(ctxTimeOut, rendezvousString, []byte(arguments[0]), dht.Quorum(1))
	if err != nil {
		fmt.Printf("Error: kademliaDHT.PutValue: %v\n", err)
		fmt.Printf(prompt)
		return
	}
	fmt.Printf(prompt)
}

func kDhtGetValue(arguments []string) {
	// todo parameterize timeout
	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(30))
	defer cancel()
	//peerId, err := peer.IDFromString(arguments[0])
	//if err != nil {
	//	fmt.Printf("Error: peer.IDFromString: %v\n", err)
	//	return
	//}
	val, err := kademliaDHT.GetValue(ctxTimeOut, rendezvousString, dht.Quorum(1))
	if err != nil {
		fmt.Printf("Error: kademliaDHT.GetValue: %v\n", err)
		fmt.Printf(prompt)
		return
	}
	fmt.Printf("%s\n", val)
	fmt.Printf(prompt)
}

func kDhtGetValues(arguments []string) {
	// todo parameterize timeout
	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*time.Duration(30))
	defer cancel()
	//peerId, err := peer.IDFromString(arguments[0])
	//if err != nil {
	//	fmt.Printf("Error: peer.IDFromString: %v\n", err)
	//	return
	//}
	recvdVals, err := kademliaDHT.GetValues(ctxTimeOut, rendezvousString, 1)
	if err != nil {
		fmt.Printf("Error: kademliaDHT.GetValue: %v\n", err)
		fmt.Printf(prompt)
		return
	}
	for i, recvdVal := range recvdVals {

		fmt.Printf("%d: %q from %v\n", i, recvdVal.Val, recvdVal.From)
	}

	fmt.Printf(prompt)
}

func kDhtClose(arguments []string) {
	err = kademliaDHT.Close()
	if err != nil {
		panic(err)
	}
	fmt.Printf(prompt)
}

//func setDebug(arguments []string) {
//
//	// Append arguments
//	last := strings.Join(arguments, " ")
//
//	if arguments[0] == "on" {
//
//	}
//
//	fmt.Printf("<CMD>: \\dht %s\n%v\t%s", last, kademliaDHT., node.ID().Pretty())
//}

//func kDht(arguments []string) {
//
//	// Append arguments
//	last := strings.Join(arguments, " ")
//
//	fmt.Printf("<CMD>: \\dht %s\n%v\t%s", last, kademliaDHT., node.ID().Pretty())
//}

func quitChat(arguments []string) {

	// Get rid of warnings
	_ = arguments

	os.Exit(0)
}
