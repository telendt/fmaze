package main

import (
	"bufio"
	"flag"
	"fmt"
	stdio "io"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/telendt/fmaze/event"
	"github.com/telendt/fmaze/io"
	"github.com/telendt/fmaze/router"
)

var nilTime time.Time

func main() {
	var (
		authTimeout       = flag.Duration("auth-timeout", 1*time.Second, "Client authentication timeout")
		clientsListenAddr = flag.String("clients-listen", ":9099", "User clients listen address")
		eventsCap         = flag.Int("events-capacity", 100000, "Capacity of unordered events store")
		flushInterval     = flag.Duration("flush-interval", 10*time.Second, "Write flush interval")
		msgBacklog        = flag.Int("msg-backlog", 10, "Client message backlog")
		noBackpressure    = flag.Bool("no-backpressure", false, "Disable client write backpressure")
		noReset           = flag.Bool("no-reset", false, "Don't reset internal state when event source disconnects")
		readBufSize       = flag.Int("read-buffer", 4096, "Read buffer size in bytes")
		sourceListenAddr  = flag.String("event-source-listen", ":9090", "Event source listen address")
		startSeq          = flag.Int64("start-sequence", 1, "Sequence start number")
		useWritev         = flag.Bool("use-writev", false, "Try to use writev instead of write syscall")
		writeBufSize      = flag.Int("write-buffer", 4096, "Write buffer size in bytes")
	)
	flag.Parse()

	rt := router.New(!*noBackpressure)

	cl, err := net.Listen("tcp", *clientsListenAddr)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		forwarder := io.NewMaxLatencyForwarder(*writeBufSize, *flushInterval, *useWritev)
		for {
			conn, err := cl.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go func() {
				defer conn.Close()
				var userID int
				conn.SetReadDeadline(time.Now().Add(*authTimeout))
				if _, err := fmt.Fscanln(conn, &userID); err != nil {
					return
				}
				c := make(chan []byte, *msgBacklog)
				unsubscribe, done, _ := rt.Subscribe(userID, c)
				defer unsubscribe()
				go func() {
					conn.SetReadDeadline(nilTime)
					_, _ = stdio.Copy(ioutil.Discard, conn)
					unsubscribe()
					close(c)
				}()
				forwarder.Forward(done, conn, c)
			}()
		}
	}()

	dispatcher := event.NewDispatcher(rt, *startSeq, *eventsCap)
	ln, err := net.Listen("tcp", *sourceListenAddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		r := bufio.NewReaderSize(conn, *readBufSize)
		for {
			line, err := r.ReadBytes('\n')
			if err != nil {
				conn.Close()
				break
			}
			event, err := event.Parse(line)
			if err != nil {
				log.Printf("%s: %q\n", err.Error(), line)
				conn.Close()
				break
			}
			if err := dispatcher.Dispatch(event); err != nil {
				log.Printf("%s: %q\n", err.Error(), line)
				conn.Close()
				break
			}
		}
		if !*noReset {
			dispatcher.Reset()
			rt.Reset()
		}
	}
}
