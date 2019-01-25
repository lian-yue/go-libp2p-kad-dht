package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"time"

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
	"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
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

	// Start with a prompt
	_, err = fmt.Fprint(os.Stdin, prompt())
	if err != nil {
		panic(err)
	}

	// Define the reader for stdin
	bio := bufio.NewReader(os.Stdin)

	// Goroutine for reading lines from stdin into a channel
	go func() {
		for {
			line, hasPrefix, err := bio.ReadLine()
			if err != nil {
				panic(err)
			}
			cmdChan <- fmt.Sprintf("%s", line)
			if hasPrefix {
				return
			}
		}
	}()

	// Goroutine for command execution
	go func() {
		for {
			select {
			case str := <-cmdChan:
				executeCommand(str)
			}
		}
	}()

	// Don't exit
	select {}
}
