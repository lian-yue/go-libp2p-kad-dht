### Hello World

The simplest scenario is to save a value and to get it afterward.

- join via provided bootstrap servers
- set a value for a key
- get the value of the key
- make it publicly available

        ./cmdtool alice
        < Jan 23 10:37:38.366 alice> join
        {<peer.ID Qm*VQKNAd> [/ip4/178.62.158.247/tcp/4001]} joined
        {<peer.ID Qm*zaDs64> [/ip4/104.236.76.40/tcp/4001]} joined
        {<peer.ID Qm*QLuvuJ> [/ip4/104.131.131.82/tcp/4001]} joined
        {<peer.ID Qm*L1KrGM> [/ip4/104.236.179.241/tcp/4001]} joined
        {<peer.ID Qm*wMKPnu> [/ip4/128.199.219.111/tcp/4001]} joined
        Joined with all 5 peers
        < Jan 23 10:37:45.987 alice> kdputkeyvalue greeting HelloWorld
        < Jan 23 10:38:26.201 alice> kdgetkeyvalue greeting
        HelloWorld
        < Jan 23 10:38:47.083 alice> advertise
        < Jan 23 10:38:50.302 alice>


Then you open another session.

- join via provided bootstrap servers
- get the value of the key

        ./cmdtool bob
        < Jan 23 10:48:02.219 bob> join
        {<peer.ID Qm*VQKNAd> [/ip4/178.62.158.247/tcp/4001]} joined
        {<peer.ID Qm*zaDs64> [/ip4/104.236.76.40/tcp/4001]} joined
        {<peer.ID Qm*QLuvuJ> [/ip4/104.131.131.82/tcp/4001]} joined
        {<peer.ID Qm*L1KrGM> [/ip4/104.236.179.241/tcp/4001]} joined
        {<peer.ID Qm*wMKPnu> [/ip4/128.199.219.111/tcp/4001]} joined
        Joined with all 5 peers
        < Jan 23 10:48:17.605 bob> kdgetkeyvalue greeting
        HelloWorld
        < Jan 23 10:49:10.927 bob>
        

### Kademlia Protocol

<table>
<tr><th>Kademlia</th><th>go-libp2p-kad-dht</th><th>./cmdtool</th></tr>
<tr><td>PING</td><td>NA</td><td>NA</td></tr>
<tr><td>STORE</td><td>func (dht *IpfsDHT) PutValue(ctx context.Context, key string, value []byte, opts ...ropts.Option) (err error)</td><td>kdputkeyvalue</td></tr>
<tr><td>FIND_NODE</td><td> </td><td> </td></tr>
<tr><td>FIND_VALUE</td><td> </td><td> </td></tr>
</table>
