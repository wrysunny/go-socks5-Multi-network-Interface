package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

var (
	srcip      string
	listenport string
)

func main() {
	//  flag parse.
	flag.StringVar(&srcip, "srcip", "", "which one ip send data?")
	flag.StringVar(&listenport, "listen", "10880", "socks5 server listen port.")
	flag.Parse()
	if srcip == "" || listenport == "" {
		flag.Usage()
		os.Exit(1)
	}
	// listen port
	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", listenport))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println(err)
		return
	}

	if buf[0] != 0x05 {
		log.Println("unsupported protocol version")
		return
	}

	numMethods := int(buf[1])
	methods := buf[2 : 2+numMethods]

	// select no authentication required
	var noAuth bool
	for _, m := range methods {
		if m == 0x00 {
			noAuth = true
			break
		}
	}

	if !noAuth {
		log.Println("no supported method found")
		return
	}

	// respond with selected method
	resp := []byte{0x05, 0x00}
	if _, err := conn.Write(resp); err != nil {
		log.Println(err)
		return
	}

	// read request
	n, err = conn.Read(buf)
	if err != nil {
		log.Println(err)
		return
	}

	if buf[0] != 0x05 {
		log.Println("unsupported protocol version")
		return
	}

	if buf[1] != 0x01 {
		log.Println("unsupported command")
		return
	}

	addrType := buf[3]
	var addr string
	var port string

	switch addrType {
	case 0x01:
		// IPv4 address
		ip := net.IP(buf[4:8])
		addr = ip.String()
	case 0x03:
		// domain name
		addrLen := int(buf[4])
		addr = string(buf[5 : 5+addrLen])
	case 0x04:
		// IPv6 address
		ip := net.IP(buf[4:20])
		addr = ip.String()
	}

	portBytes := buf[n-2 : n]
	port = fmt.Sprintf("%d", binary.BigEndian.Uint16(portBytes))

	// 指定ip
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		LocalAddr: &net.TCPAddr{
			IP: net.ParseIP(srcip),
			//Port: 0,
		},
	}
	target, err := dialer.DialContext(context.Background(), "tcp", net.JoinHostPort(addr, port))
	//target, err := dialer.Dial("tcp", net.JoinHostPort(addr, port))
	//target, err := net.Dial("tcp", net.JoinHostPort(addr, port))
	if err != nil {
		log.Println(err)
		return
	}
	defer target.Close()

	// respond with success
	resp = []byte{0x05, 0x00, 0x00, addrType}
	if addrType == 0x01 {
		ip := net.ParseIP(addr).To4()
		resp = append(resp, ip...)
	} else if addrType == 0x03 {
		resp = append(resp, byte(len(addr)))
		resp = append(resp, []byte(addr)...)
	} else if addrType == 0x04 {
		ip := net.ParseIP(addr).To16()
		resp = append(resp, ip...)
	}

	portNum := make([]byte, 2)
	binary.BigEndian.PutUint16(portNum, uint16(target.LocalAddr().(*net.TCPAddr).Port))
	resp = append(resp, portNum...)

	if _, err := conn.Write(resp); err != nil {
		log.Println(err)
		return
	}

	// relay traffic between client and target
	go func() {
		_, err := io.Copy(conn, target)
		if err != nil {
			log.Println(err)
		}
	}()

	_, err = io.Copy(target, conn)
	if err != nil {
		log.Println(err)
	}
}
