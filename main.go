package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

func handleClients(ln net.Listener, s Subscriber) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			var userID int
			if _, err := fmt.Fscanln(conn, &userID); err != nil {
				conn.Close()
				return
			}
			c := make(chan []byte, 10) // TODO: parametrize
			unsubscribe, _ := s.Subscribe(userID, c)
			defer unsubscribe()
			go func() {
				_, _ = io.Copy(ioutil.Discard, conn)
				unsubscribe()
				close(c)
			}()
			w := bufio.NewWriter(conn) // TODO: parametrize
			for msg := range c {
				if _, err := w.Write(msg); err != nil {
					return
				}
			loop:
				for { // try to get more
					select {
					case msg, more := <-c:
						if !more {
							return
						}
						if _, err := w.Write(msg); err != nil {
							return
						}
					default:
						break loop
					}
				}
				if err := w.Flush(); err != nil {
					return
				}
			}
		}()
	}
}

func main() {
	var (
		userClientsListenAddr = flag.String("user-clients-listen", ":9099", "User clients listen address")
		eventSourceListenAddr = flag.String("eventsource-listen", ":9090", "Event source listen address")
		eventsCap             = flag.Int("events-capacity", 100000, "Capacity of an array with unordered events")
		backpressure          = flag.Bool("clients-backpressure", false, "Enable client write backpressure")
	)
	flag.Parse()

	userGraph := NewUserGraph(*backpressure)

	cl, err := net.Listen("tcp", *userClientsListenAddr)
	if err != nil {
		log.Fatal(err)
	}

	go handleClients(cl, userGraph)

	dispatcher := NewDispatcher(userGraph, 1, *eventsCap)
	ln, err := net.Listen("tcp", *eventSourceListenAddr)
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
