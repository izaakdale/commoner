package udpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/izaakdale/commoner/internal/udpserver"
	"golang.org/x/time/rate"
)

func Run(ctx context.Context, addr string, hz, burst int) error {
	lsAddrPort, err := netip.ParseAddrPort(addr)
	if err != nil {
		log.Fatal(err)
	}
REDIAL:
	udpLsAddr := net.UDPAddrFromAddrPort(lsAddrPort)
	dial, err := net.DialUDP("udp", nil, udpLsAddr)
	if err != nil {
		log.Printf("error dialing: %s\n", err.Error())
		goto REDIAL
	}

	name := uuid.NewString()

	enc := json.NewEncoder(dial)
	lm := rate.NewLimiter(rate.Every(time.Second/time.Duration(hz)), burst)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := lm.Wait(ctx); err != nil {
				return fmt.Errorf("error waiting: %w", err)
			}
			out := udpserver.RemoteServer{
				Name: name,
				Addr: dial.LocalAddr().String(),
			}
			if err := enc.Encode(out); err != nil {
				oe := &net.OpError{}
				if errors.As(err, &oe) {
					goto REDIAL
				}
				log.Printf("error encoding message: %s", err.Error())
				continue
			}
		}
	}
}
