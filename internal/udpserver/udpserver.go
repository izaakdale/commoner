package udpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

type RemoteServer struct {
	Name       string    `json:"name"`
	Addr       string    `json:"addr"`
	LastActive time.Time `json:"last_active"`
}

var (
	mu    = new(sync.Mutex)
	addrs = make(map[string]RemoteServer)
)

func Run(ctx context.Context, conn net.Conn) error {
	for {
		var received RemoteServer
		err := json.NewDecoder(conn).Decode(&received)
		if err != nil {
			fmt.Println("Decoding error:", err)
			continue
		}
		mu.Lock()
		received.LastActive = time.Now()
		addrs[received.Name] = received
		mu.Unlock()
	}
}

func GetRemoteServers(timeout time.Duration) []RemoteServer {
	var servers []RemoteServer
	mu.Lock()
	for _, addr := range addrs {
		if time.Since(addr.LastActive) < timeout {
			servers = append(servers, addr)
		} else {
			delete(addrs, addr.Name)
		}
	}
	mu.Unlock()
	return servers
}
