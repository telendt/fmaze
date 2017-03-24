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
			_, err := fmt.Fscanln(conn, &userID)
			if err != nil {
				conn.Close()
			}
			c := make(chan []byte, 10) // TODO: parametrize
			unsubscribe, err := s.Subscribe(userID, c)
			done := make(chan struct{})
			go func() {
				_, _ = io.Copy(ioutil.Discard, conn)
				close(done)
			}()
			w := bufio.NewWriter(conn) // TODO: parametrize
		outer:
			for {
				select {
				case <-done:
					break outer
				case msg := <-c:
					if _, err := w.Write(msg); err != nil {
						break outer
					}
				inner:
					for {
						select {
						case msg := <-c:
							if _, err := w.Write(msg); err != nil {
								break outer
							}
						default:
							if err := w.Flush(); err != nil {
								break outer
							}
							break inner
						}
					}
				}
			}
			unsubscribe()
		}()
	}
}

func main() {
	var (
		clientListener = flag.String("clients-listener", ":9099", "User clients listener")
		eventListener  = flag.String("event-listener", ":9090", "Event source listener")
		eventsCap      = flag.Int("events-capacity", 100000, "Capacity of an array with unordered events")
		backpressure   = flag.Bool("clients-backpresure", false, "Enable client write backpressure")
	)
	flag.Parse()

	userGraph := NewUserGraph(*backpressure)

	cl, err := net.Listen("tcp", *clientListener)
	if err != nil {
		log.Fatal(err)
	}

	go handleClients(cl, userGraph)

	dispatcher := NewDispatcher(userGraph, 1, *eventsCap)
	ln, err := net.Listen("tcp", *eventListener)
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
			dispatcher.Dispatch(event)
		}
		dispatcher.Reset()
		userGraph.Reset()
	}
}
