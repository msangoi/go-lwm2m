package main

import (
	"log"
	"github.com/msangoi/go-coap"
	"os"
	"net"
	"time"
	"math/rand"
)

func main() {

	if len(os.Args) != 3 {
		log.Fatalf("Invalid number of arguments, expected <serverHost> <serverPort>")
		os.Exit(-1)
	}
	serverHost := os.Args[1]
	serverPort := os.Args[2]

	log.Println("Starting LWM2M client")

	rand.Seed(time.Now().Unix())
	msgId := rand.Intn(10000)

	// binding to a UDP socket
	laddr, err := net.ResolveUDPAddr("udp", ":5685")
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	c, err := net.ListenUDP("udp", laddr)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	log.Printf("connection: %v", c)

	// Send register request
	register := coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.POST,
		MessageID: uint16(msgId),
		Payload:   []byte("</3/0>"),
	}
	register.SetPathString("/rd")
	register.AddOption(coap.URIQuery, "ep=goclient");	

	// server address
	uaddr, err := net.ResolveUDPAddr("udp", serverHost + ":" + serverPort)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
	err = coap.Transmit(c, uaddr, register)
	if err != nil {
		log.Fatalf("Error while sending registration request: %v", err)
		os.Exit(-1)
	}
	buf := make([]byte, 1500)
	rv, err := coap.Receive(c, buf)

	if err != nil {
		log.Fatalf("Error while sending registration request: %v", err)
		os.Exit(-1)
	}

	if &rv != nil {
		log.Printf("Ack received: %v", &rv)
		log.Printf("Registered with id: %s", rv.Options(coap.LocationPath)[1])
	}	

	// Listen for incoming requests
	rh := coap.FuncHandler(func(l *net.UDPConn, a *net.UDPAddr, m *coap.Message) *coap.Message {
			log.Printf("New message: path=%q: %#v from %v", m.Path(), m, a)
			if m.IsConfirmable() {
				res := &coap.Message{
					Type:      coap.Acknowledgement,
					Code:      coap.Content,
					MessageID: m.MessageID,
					Token:     m.Token,
					Payload:   []byte("Plop"),
				}

				res.SetOption(coap.ContentFormat, coap.TextPlain)
				for _, p := range m.Path() {
					m.AddOption(coap.LocationPath, p)
				}

				return res
			}
			 return nil
		})

	c.SetReadDeadline(time.Time{})
	buf = make([]byte, 1500)
	for {
		nr, addr, err := c.ReadFromUDP(buf)
		if err == nil {			
			tmp := make([]byte, nr)
			copy(tmp, buf)
			go coap.HandlePacket(c, tmp, addr, rh)
		}
	}

}