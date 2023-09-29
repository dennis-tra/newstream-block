package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func newHost(t *testing.T) host.Host {
	listenAddr := libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0")

	h, err := libp2p.New(listenAddr)
	require.NoError(t, err)

	t.Cleanup(func() {
		if err = h.Close(); err != nil {
			t.Logf("unexpected error when closing host: %s", err)
		}
	})
	return h
}

func TestListenCloseCount3(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Run(fmt.Sprintf("[%d]", i), func(t *testing.T) {
			h1 := newHost(t)
			h2 := newHost(t)
			fmt.Println(time.Now().Format(time.RFC3339Nano)+" H1:", h1.ID())
			fmt.Println(time.Now().Format(time.RFC3339Nano)+" H2:", h2.ID())

			h1.Network().Notify(&network.NotifyBundle{
				ListenF: func(n network.Network, multiaddr ma.Multiaddr) {
					fmt.Println(time.Now().Format(time.RFC3339Nano)+" H1 started listening on:", multiaddr)
				},
				ConnectedF: func(n network.Network, conn network.Conn) {
					if conn.RemotePeer() != h2.ID() {
						panic("aahh")
					}
					fmt.Println(time.Now().Format(time.RFC3339Nano) + " H1 connected to: H2")
				},
			})

			h2.Network().Notify(&network.NotifyBundle{
				ListenF: func(n network.Network, multiaddr ma.Multiaddr) {
					fmt.Println(time.Now().Format(time.RFC3339Nano)+" H2 started listening on:", multiaddr)
				},
				ConnectedF: func(n network.Network, conn network.Conn) {
					if conn.RemotePeer() != h1.ID() {
						panic("aahh")
					}
					fmt.Println(time.Now().Format(time.RFC3339Nano) + " H2 connected to: H1")
				},
			})

			h1.Peerstore().AddAddrs(h2.ID(), h2.Addrs(), peerstore.PermanentAddrTTL)
			h2.Peerstore().AddAddrs(h1.ID(), h1.Addrs(), peerstore.PermanentAddrTTL)

			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				fmt.Println(time.Now().Format(time.RFC3339Nano) + " H1 -> H2 NewStream")
				_, _ = h1.NewStream(context.Background(), h2.ID(), "/any/protocol")
				fmt.Println(time.Now().Format(time.RFC3339Nano) + " H1 -> H2 NewStream Done!")
			}()

			go func() {
				defer wg.Done()
				fmt.Println(time.Now().Format(time.RFC3339Nano) + " H2 -> H1 NewStream")
				_, _ = h2.NewStream(context.Background(), h1.ID(), "/any/protocol")
				fmt.Println(time.Now().Format(time.RFC3339Nano) + " H2 -> H1 NewStream Done!")
			}()
			wg.Wait()
		})
	}
}
