package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
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

func TestBlock(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Run(fmt.Sprintf("[%d]", i), func(t *testing.T) {
			h1 := newHost(t)
			h2 := newHost(t)

			h1.Peerstore().AddAddrs(h2.ID(), h2.Addrs(), peerstore.PermanentAddrTTL)
			h2.Peerstore().AddAddrs(h1.ID(), h1.Addrs(), peerstore.PermanentAddrTTL)

			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				_, _ = h1.NewStream(context.Background(), h2.ID(), "/any/protocol")
			}()

			go func() {
				defer wg.Done()
				_, _ = h2.NewStream(context.Background(), h1.ID(), "/any/protocol")
			}()

			wg.Wait()
		})
	}
}
