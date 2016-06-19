package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

// Config for hyperon server
type Config struct {
	LocalAddr  string
	RemoteAddr string
}

// NewConfigFromFile will load config from toml style file
func NewConfigFromFile(path string) (*Config, error) {
	cfg := Config{}
	_, err := toml.DecodeFile(path, &cfg)

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func transfer(client *net.Conn, server *net.Conn) {
	var c = sync.NewCond(&sync.Mutex{})
	var done = false

	go func() {
		io.Copy(*server, *client)
		c.L.Lock()
		done = true
		c.Signal()
		c.L.Unlock()
	}()
	go func() {
		io.Copy(*client, *server)
		c.L.Lock()
		done = true
		c.Signal()
		c.L.Unlock()
	}()
	c.L.Lock()
	for !done {
		c.Wait()
	}
	c.L.Unlock()
}

func bridge(local string, remote string) {
	ln, err := net.Listen("tcp", local)
	defer ln.Close()

	if err != nil {
		log.Printf("fail to listen on %s, %s", local, err)
		return
	}
	log.Printf("listen on %s", local)
	for {
		client, err := ln.Accept()
		if err != nil {
			log.Printf("fail to accept on %s, %s", local, err)
		}
		tc, ok := client.(*net.TCPConn)
		if !ok {
			log.Printf("error to cast connection")
		}
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(1 * time.Second)
		defer client.Close()

		server, err := net.Dial("tcp", remote)
		if err != nil {
			log.Printf("fail to connect to server %s, %s", remote, err)
		}
		defer server.Close()
		go func() {
			log.Printf("establish %s --> %s", local, remote)
			transfer(&client, &server)
			log.Printf("closing %s --> %s", local, remote)
		}()
	}
}

func main() {
	path := "hyperon.toml"
	cfg, err := NewConfigFromFile(path)
	if err != nil {
		fmt.Printf("unable load config from %s, %s\n", path, err)
	}
	bridge(cfg.LocalAddr, cfg.RemoteAddr)
	select {}
}
