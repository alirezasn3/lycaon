package main

import (
	"context"
	"encoding/json"
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

type App struct {
	ctx context.Context
}

type Hop struct {
	Number    int      `json:"number"`
	Address   string   `json:"address"`
	IPEEInfo  IPEEInfo `json:"ipeeInfo"`
	RTT       int64    `json:"rtt"`
	IsPrivate bool     `json:"isPrivate"`
}

type IPEEInfo struct {
	OK               bool   `json:"ok"`
	Type             string `json:"type"`
	CIDR             string `json:"cidr"`
	ASNumber         int    `json:"asNumber"`
	BinarySubnetMask string `json:"binarySubnetMask"`
	SubnetMask       string `json:"subnetMask"`
	Class            string `json:"class"`
	NetworkAddress   string `json:"networkAddress"`
	NumberOfHosts    int    `json:"numberOfHosts"`
	ASName           string `json:"asName"`
	OrganizationName string `json:"organizationName"`
	Country          string `json:"country"`
	CountryCode      string `json:"countryCode"`
	QueryDuration    int    `json:"t"`
	Query            string `json:"query"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// trace route
func (a *App) Trace(ip string, maxHops int, timeout int) string {
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

	// rtt timer
	var start int64

	// buffer
	b := make([]byte, 8192)

	// range over 1 to max ttl
	for i := 1; i <= maxHops; i++ {
		// set ttl
		if err := listener.IPv4PacketConn().SetTTL(i); err != nil {
			return err.Error()
		}

		// set start time
		start = time.Now().UnixMilli()

		// send packet
		if _, err := listener.WriteTo(icmpMessageBytes, &net.IPAddr{IP: net.ParseIP(ip)}); err != nil {
			return err.Error()
		}

		// set read deadline
		if err = listener.IPv4PacketConn().SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(timeout))); err != nil {
			return err.Error()
		}

		// wait for response
		_, src, err := listener.ReadFrom(b)
		if err != nil {
			runtime.EventsEmit(a.ctx, "hop", Hop{Number: i, Address: "timeout"})
			continue
		}

		hop := &Hop{Number: i, Address: src.String(), RTT: time.Now().UnixMilli() - start, IsPrivate: net.ParseIP(src.String()).IsPrivate()}

		// send hop details to frontend
		runtime.EventsEmit(a.ctx, "hop", hop)

		// get IPEEInfo if ip is not private
		if !hop.IsPrivate {
			// increment wait group
			wg.Add(1)

			// send request to api on new thread
			go func(h *Hop) {
				res, err := http.Get("https://ipee-api.alirezasn.workers.dev/v1/info/" + h.Address)
				if err != nil {
					log.Println(err)
					return
				}
				bytes, err := io.ReadAll(res.Body)
				if err != nil {
					log.Println(err)
					return
				}
				res.Body.Close()

				var info IPEEInfo
				err = json.Unmarshal(bytes, &info)
				if err != nil {
					log.Println(err)
					return
				}

				// send info to frontend
				h.IPEEInfo = info
				runtime.EventsEmit(a.ctx, "hop info", h)

				// decrement wait group
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

	return ""
}

// get app version from embeded wails.json file
func (a *App) GetVersion() string {
	var config WailsConfig
	err := json.Unmarshal(wailsConfigBytes, &config)
	if err != nil {
		log.Println(err)
		return "unknown"
	}
	return config.Info.ProductVersion
}
