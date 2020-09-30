package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/miekg/dns"
)

const eth10Addr = "10.0.0.1"
const eth11Addr = "10.0.0.2"
const ttl = 0

var closeConn = flag.Bool("close-idle-conns", false, "causes the http client to close idle conns")

var records = map[string]string{
	"foo.service.": eth10Addr,
}

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			log.Println("[DNS] received query for " + q.Name)
			ip := records[q.Name]
			if ip != "" {
				rr := handleARecord(q.Name)
				m.Answer = append(m.Answer, rr)
			}
		}
	}
}

func handleARecord(recordName string) dns.RR {
	rr, err := dns.NewRR(fmt.Sprintf("%s A %s", recordName, records[recordName]))
	if err == nil {
		rr.Header().Ttl = ttl
	} else {
		log.Println("[DNS] failed to serve " + recordName + ": " + err.Error())
	}
	return rr
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	if r.Opcode == dns.OpcodeQuery {
		parseQuery(m)
	}

	_ = w.WriteMsg(m)
}

func main() {
	flag.Parse()

	// attach request handler func
	dns.HandleFunc("service.", handleDnsRequest)

	// start server
	port := 8053
	server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
	log.Printf("[DNS] starting on :%d\n", port)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start server: %s\n", err.Error())
		}
	}()
	serveHTTP("ETH10", eth10Addr)
	serveHTTP("ETH11", eth11Addr)
	queryDNS()
}

func serveHTTP(iface, addr string) {
	server := &http.Server{Addr: addr + ":8080"}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("hello on " + addr + ":8080!"))
	})
	server.Handler = mux

	go func() {
		log.Printf("[%s] starting on %s:8080\n", iface, addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[%s] failed to start HTTP server on %s:8080: %s\n", iface, addr, err.Error())
		}
	}()
}

func queryDNS() {
	time.Sleep(time.Second)

	// set PreferGo in order to override the Dial func
	net.DefaultResolver.PreferGo = true
	net.DefaultResolver.Dial = func(ctx context.Context, network, addr string) (net.Conn, error) {
		d := net.Dialer{}
		if network == "udp" {
			return d.DialContext(ctx, "udp", ":8053")
		}
		return d.DialContext(ctx, network, addr)
	}

	msg1 := doRequest("http://foo.service.:8080")
	log.Printf("[HTTP] request response 1: %s", msg1)

	// update foo.service. record to point to eth11
	records = map[string]string{
		"foo.service.": eth11Addr,
	}

	if *closeConn {
		http.DefaultClient.CloseIdleConnections()
	}

	msg2 := doRequest("http://foo.service.:8080")
	log.Printf("[HTTP] request response 2: %s", msg2)

	// check if we got a response from eth11 or eth10
	if msg2 == "hello on "+eth10Addr+":8080!" {
		fmt.Printf("received message from %s but expected %s\n", eth10Addr, eth11Addr)
		os.Exit(1)
	} else if msg2 == "hello on "+eth11Addr+":8080!" {
		os.Exit(0)
	}
	log.Fatalf("unknown response: " + msg2)
}

func doRequest(location string) string {
	resp, err := http.Get(location)
	if err != nil {
		log.Fatalf("[HTTP] failed to do request to %s: %s\n", location, err.Error())
	}
	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("[HTTP] failed to read request body:%s\n", err.Error())
	}
	_ = resp.Body.Close()
	return string(msg)
}
