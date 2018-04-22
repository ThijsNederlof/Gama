package state

import (
	"../model"
	"fmt"
	"sync/atomic"
)

var (
	Clients  = make(map[uint64]*model.NetClient)
	clientId uint64
)

// Adds client to active clients and returns player id
func AddClient(client *model.NetClient) uint64 {
	clientNumber := atomic.AddUint64(&clientId, 1)

	Clients[clientNumber] = client

	return clientNumber
}

func GetClient(clientId uint64) *model.NetClient {
	return Clients[clientId]
}

func RemoveClient(clientId uint64) {
	fmt.Println("Removing client from state")
	delete(Clients, clientId)
}

func GetClientList() map[uint64]*model.NetClient {
	return Clients
}
