package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"
)

func main() {
	var (
		clientsListenAddr = flag.String("clients-listen", ":9099", "User clients listen address")
		sourceListenAddr  = flag.String("event-source-listen", ":9090", "Event source listen address")
		eventsCap         = flag.Int("events-capacity", 100000, "Capacity of unordered events store")
		noBackpressure    = flag.Bool("no-backpressure", false, "Disable client write backpressure")
		writeBufSize      = flag.Int("write-buffer", 4096, "Write buffer size in bytes")
		useWritev         = flag.Bool("use-writev", false, "Try to use writev instead of write syscall")
		msgBacklog        = flag.Int("msg-backlog", 10, "Client message backlog")
		flushInterval     = flag.Duration("flush-interval", 10*time.Second, "Write flush interval")
	)
	flag.Parse()

	userGraph := NewUserGraph(!*noBackpressure)

	cl, err := net.Listen("tcp", *clientsListenAddr)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		forwarder := NewMaxLatencyForwarder(*writeBufSize, *flushInterval, *useWritev)
		for {
			conn, err := cl.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go func() {
				var userID int
				if _, err := fmt.Fscanln(conn, &userID); err != nil {
					return
				}
				c := make(chan []byte, *msgBacklog)
				unsubscribe, done, _ := userGraph.Subscribe(userID, c)
				defer unsubscribe()
				go func() {
					_, _ = io.Copy(ioutil.Discard, conn)
					unsubscribe()
					close(c)
				}()
				forwarder.Forward(done, conn, c)
				conn.Close()
			}()
		}
	}()

	dispatcher := NewDispatcher(userGraph, 1, *eventsCap)
	ln, err := net.Listen("tcp", *sourceListenAddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		r := bufio.NewReader(conn)
		for {
			line, err := r.ReadBytes('\n')
			if err != nil {
				conn.Close()
				break
			}
			event, err := ParseEvent(line)
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
		dispatcher.Reset()
		userGraph.Reset()
	}
}
