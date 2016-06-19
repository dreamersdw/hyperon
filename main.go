package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

// Config for hyperon server
type ServerConf struct {
	name       string
	LocalAddr  string
	RemoteAddr string
}
type Config struct {
	Servers []ServerConf
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

func handle(client net.Conn, local string, remote string) {
	tc, ok := client.(*net.TCPConn)
	if !ok {
		log.Printf("error to cast connection")
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(1 * time.Second)
	defer client.Close()

	server, err := net.Dial("tcp", remote)
	if err != nil {
		log.Printf("fail to connect to server %s, %s", remote, err)
		return
	}
	defer server.Close()

	log.Printf("establish %s --> %s", local, remote)
	transfer(&client, &server)
	log.Printf("closing %s --> %s", local, remote)
}

func bridge(local string, remote string) {
	ln, err := net.Listen("tcp", local)

	if err != nil {
		log.Printf("%s", err)
		return
	}
	defer ln.Close()

	log.Printf("listen on %s", local)
	for {
		client, err := ln.Accept()
		if err != nil {
			log.Printf("fail to accept on %s, %s", local, err)
			continue
		}
		go handle(client, local, remote)
	}
}

func main() {
	conf := flag.String("c", "hyperon.toml", "hyperon config file")
	flag.Parse()
	path := *conf

	cfg, err := NewConfigFromFile(path)
	if err != nil {
		fmt.Printf("unable load config from %s, %s\n", path, err)
	}

	for _, serverconf := range cfg.Servers {
		localAddr := serverconf.LocalAddr
		remoteAddr := serverconf.RemoteAddr
		go bridge(localAddr, remoteAddr)
	}
	select {}
}
