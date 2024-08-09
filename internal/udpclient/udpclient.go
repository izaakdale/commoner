package udpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/izaakdale/commoner/internal/udpserver"
	"golang.org/x/time/rate"
)

func Run(ctx context.Context, addr string, hz, burst int) error {

	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Printf("error dialing: %s\n", err.Error())
		return Run(ctx, addr, hz, burst)
	}

	name := uuid.NewString()

	enc := json.NewEncoder(conn)
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
				Addr: conn.LocalAddr().String(),
			}
			if err := enc.Encode(out); err != nil {
				oe := &net.OpError{}
				if errors.As(err, &oe) {
					return Run(ctx, addr, hz, burst)
				}
				log.Printf("error encoding message: %s", err.Error())
				continue
			}
		}
	}
}
