package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/peterh/liner"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-kad-dht/opts"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
	"github.com/stefanhans/go-libp2p-kad-dht"
)

var (
	name string

	err         error
	ctx         = context.Background()
	node        host.Host
	kademliaDHT *dht.IpfsDHT

	bootstrapPeerAddrs []multiaddr.Multiaddr

	rendezvousString string
	rendezvousPoint  cid.Cid
	v1b              = cid.V1Builder{Codec: cid.Raw, MhType: multihash.SHA2_256}

	routingDiscovery *discovery.RoutingDiscovery
)

// IPFS bootstrap nodes. Used to find other peers in the network.
var bootstrapPeers = []string{
	//"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	//"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	//"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	//"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	//"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
	"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",            // mars.i.ipfs.io
	"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",           // pluto.i.ipfs.io
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",           // saturn.i.ipfs.io
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",             // venus.i.ipfs.io
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",            // earth.i.ipfs.io
	"/ip6/2604:a880:1:20::203:d001/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",  // pluto.i.ipfs.io
	"/ip6/2400:6180:0:d0::151:6001/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",  // saturn.i.ipfs.io
	"/ip6/2604:a880:800:10::4a:5001/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64", // venus.i.ipfs.io
	"/ip6/2a03:b0c0:0:1010::23:1001/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd", // earth.i.ipfs.io
}

func prompt() string {
	return fmt.Sprintf("< %s %s> ", time.Now().Format("Jan 2 15:04:05.000"), name)
}

func main() {

	// debug switches on debugging
	debug := flag.Bool("debug", false, "switches on debugging")

	// debugfilename is the file to write debugging output
	debugfilename := flag.String("debugfile", "", "file to write debugging output")

	// Parse input and check arguments
	flag.Parse()
	if flag.NArg() < 1 {
		_, _ = fmt.Fprintln(os.Stderr, "missing or wrong parameter: <name>")
		// todo usage
		os.Exit(1)
	}
	name = flag.Arg(0)

	// Start debugging to file, if switched on
	if *debug {
		debugfile, err := startLogging(*debugfilename)
		if err != nil {
			panic(err)
		}
		defer debugfile.Close()
	}

	// libp2p.New constructs a new libp2p Host. Other options can be added
	// here.
	node, err = libp2p.New(ctx, libp2p.DisableRelay())
	//libp2p.Identity(privRedId),
	//libp2p.ListenAddrs([]multiaddr.Multiaddr(config.ListenAddresses)...),
	if err != nil {
		panic(err)
	}

	//s := liner.NewLiner()
	//s.SetTabCompletionStyle(liner.TabPrints)
	//s.SetCompleter(func(line string) (ret []string) {
	//	for _, c := range commandKeys {
	//		if strings.HasPrefix(c, line) {
	//			ret = append(ret, c)
	//		}
	//	}
	//	return
	//})
	//defer s.Close()

	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err = dht.New(ctx, node, dhtopts.Validator(NullValidator{}))
	if err != nil {
		panic(err)
	}

	// Initialize commands
	commandsInit()

	// Start loop with history and completion
	interactiveLoop(kademliaDHT, node)
}

func interactiveLoop(d *dht.IpfsDHT, h host.Host) error {
	s := liner.NewLiner()
	s.SetTabCompletionStyle(liner.TabPrints)
	s.SetCompleter(func(line string) (ret []string) {
		for _, c := range commandKeys {
			if strings.HasPrefix(c, line) {
				ret = append(ret, c)
			}
		}
		return
	})
	defer s.Close()
	for {
		p, err := s.Prompt(prompt())
		if err == io.EOF {
			return nil
		}
		if err != nil {
			panic(err)
		}
		if executeCommand(p) {
			s.AppendHistory(p)
		}
	}
}
