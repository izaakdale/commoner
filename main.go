package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"time"

	"github.com/izaakdale/commoner/internal/udpclient"
	"github.com/izaakdale/commoner/internal/udpserver"
	"github.com/kelseyhightower/envconfig"
)

type specification struct {
	API_ADDR          string `envconfig:"API_ADDR" required:"true"`
	UDPDialAddr       string `envconfig:"UDP_DIAL_ADDR" required:"true"`
	UDPListenAddr     string `envconfig:"UDP_LISTEN_ADDR" required:"true"`
	BROADCAST_FREQ_HZ int    `envconfig:"BROADCAST_FREQ_HZ" required:"true"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	var spec specification
	if err := envconfig.Process("APP", &spec); err != nil {
		log.Fatal(err)
	}

	errCh := make(chan error, 1)

	addrPort, err := netip.ParseAddrPort(spec.UDPListenAddr)
	if err != nil {
		log.Fatal(err)
	}
	udpAddr := net.UDPAddrFromAddrPort(addrPort)
	ls, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	go func(ec chan error) { ec <- udpserver.Run(ctx, ls) }(errCh)

	// Ideally wouldn't sleep here
	go func(ec chan error) {
		ec <- udpclient.Run(ctx, spec.UDPDialAddr, spec.BROADCAST_FREQ_HZ, 1)
	}(errCh)

	mux := http.NewServeMux()
	mux.HandleFunc("/servers", func(w http.ResponseWriter, r *http.Request) {
		// to ensure that the servers are up to date, timeout is set to 1/10th of the broadcast frequency
		// if timeout and broadcast frequency are the same, then it is possible that time.Since is greater than the timeout.
		srvs := udpserver.GetRemoteServers(time.Second / (time.Duration(spec.BROADCAST_FREQ_HZ) / 10))
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(srvs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	go func(ec chan error) {
		ec <- http.ListenAndServe(spec.API_ADDR, mux)
	}(errCh)

	log.Printf("successfully connected\n")
	select {
	case err := <-errCh:
		log.Fatal(err)
	case <-ctx.Done():
		if ctx.Err() != context.Canceled {
			log.Fatal(ctx.Err())
		}
		cancel()
		fmt.Println("Exiting")
	}
}
