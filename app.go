package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// App struct
type App struct {
	ctx context.Context
}

type Hop struct {
	Number   int      `json:"number"`
	Address  string   `json:"address"`
	IPEEInfo IPEEInfo `json:"ipeeInfo"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) Trace(ip string) string {
	// slice to store each hop
	var hops []*Hop

	// icmp message
	icmpMessage := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:  os.Getpid() & 0xffff,
			Seq: 1,
		},
	}

	// marshal icmp message into bytes
	icmpMessageBytes, err := icmpMessage.Marshal(nil)
	if err != nil {
		return err.Error()
	}

	// create icmp packet listener
	listener, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return err.Error()
	}
	defer listener.Close()

	// create wait group
	var wg sync.WaitGroup

	// range over 1 to max ttl
	for i := 1; i < 32; i++ {
		// set ttl
		if err := listener.IPv4PacketConn().SetTTL(i); err != nil {
			return err.Error()
		}

		// send packet
		if _, err := listener.WriteTo(icmpMessageBytes, &net.IPAddr{IP: net.ParseIP(ip)}); err != nil {
			return err.Error()
		}

		// set read deadline
		if err = listener.IPv4PacketConn().SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			return err.Error()
		}

		// wait for response
		b := make([]byte, 1024)
		_, src, err := listener.ReadFrom(b)
		if err != nil {
			hops = append(hops, &Hop{Number: i, Address: "timeout"})
			runtime.EventsEmit(a.ctx, "hop", Hop{Number: i, Address: "timeout"})
			continue
		}

		hop := &Hop{Number: i, Address: src.String()}

		// save response
		hops = append(hops, hop)
		runtime.EventsEmit(a.ctx, "hop", Hop{Number: i, Address: src.String()})

		if !net.ParseIP(src.String()).IsPrivate() {
			wg.Add(1)
			go func(h *Hop) {
				res, err := http.Get("https://api.ipee.info/v1/info/" + h.Address)
				if err != nil {
					log.Println(err)
				}
				bytes, err := io.ReadAll(res.Body)
				if err != nil {
					log.Println(err)
				}
				res.Body.Close()
				var info IPEEInfo
				err = json.Unmarshal(bytes, &info)
				if err != nil {
					log.Println(err)
				}
				h.IPEEInfo = info
				runtime.EventsEmit(a.ctx, "hop info", h)
				wg.Done()
			}(hop)
		}

		// break the loop if destination is reached
		if src != nil && src.String() == ip {
			break
		}
	}

	// wait for go routines to finish
	wg.Wait()
	for _, hop := range hops {
		fmt.Println(hop.Number, hop.Address, hop.IPEEInfo.Country, hop.IPEEInfo.ASName)
	}

	return ""
}
