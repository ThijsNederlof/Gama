package network

import (
	"../event"
	"../model"
	"../state"
	"fmt"
	proto "github.com/golang/protobuf/proto"
	"net"
	"os"
	"strconv"
	"time"
)

var (
	newConns  = make(chan net.Conn, 128)
	deadConns = make(chan net.Conn, 128)
)

func StartTCPServer(port int) {
	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("Error starting server:", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Listening on port", port)

	go connectionHandler()
	go connectionManager()

	for {
		conn, err := listener.Accept()
		if err != nil {
			os.Exit(1)
		}
		newConns <- conn
	}
}

func connectionHandler() {
	for {
		select {
		case conn := <-newConns:
			go func(conn net.Conn) {
				initClient(conn)
			}(conn)

		case deadConns := <-deadConns:
			deadConns.Close()
		}
	}
}

func connectionManager() {
	for {
		for k, v := range state.Clients {
			d := time.Since(v.LastMessageTime)
			fmt.Println(d)
			if d < (2 * time.Second) {
				continue
			}

			if d > (2*time.Second) && d < (10*time.Second) {
				fmt.Println("longer than 2 seconds")
				h := &events.Heartbeat{
					Message: "hello!",
					Id:      1,
				}

				out, err := proto.Marshal(h)
				if err != nil {
					fmt.Println("Error parsing proto message")
				}

				v.Out <- out
				continue
			}

			fmt.Println("longer than 10 seconds")
			v.CloseChannels()
			deadConns <- v.Conn
			state.RemoveClient(k)
		}

		time.Sleep(2 * time.Second)
	}
}

func initClient(conn net.Conn) {
	client := model.NetClient{
		LastMessageTime: time.Now(),
		Out:             make(chan []byte, 25),
		In:              make(chan []byte, 25),
		Conn:            conn,
	}

	//Add client to active clients
	clientId := state.AddClient(&client)
	fmt.Println("New client #", clientId, "connected")

	// Receive on new goroutine
	go func() {
		buf := make([]byte, 1024)
		for {
			nbyte, err := conn.Read(buf)
			if err != nil {
				fmt.Println("Client #", clientId, "disconnected", err.Error())
				deadConns <- conn
				state.RemoveClient(clientId)
				break
			}

			client.LastMessageTime = time.Now()
			fragment := make([]byte, nbyte)
			copy(fragment, buf[:nbyte])
			client.In <- fragment
		}
	}()

	// SendString on new goroutine
	go func() {
		for b := range client.Out {
			fmt.Println("writing message", b)
			conn.Write(b)
		}
	}()

	// Execute on new goroutine
	go func() {
		for m := range client.In {
			h := &events.Heartbeat{}

			proto.Unmarshal(m, h)
			// Process message here
			fmt.Println("Message from user:", clientId, "with message:", h.Id, h.Message)
		}
	}()
}
