package websocket

import (
	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients 	map[string]map[*Client]struct{}
	// Broadcast to client
	broadcast 	chan []byte
	// Register requests from the clients.
	register 	chan *Client
	// Unregister requests from clients.
	unregister 	chan *Client
	//Message received from clients
	received	chan []byte
	//If isPrivate == true, broadcast will check client ID
	isPrivate	bool
}

type Message struct {
	ID		string		`json:"id"`
	Data	interface{}	`json:"data"`
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			if _, exist := h.clients[client.ID]; !exist {
				h.clients[client.ID] = make(map[*Client]struct{})
			}
			h.clients[client.ID][client] = struct{}{}
			go client.writePump()
			go client.readPump()
		case client := <-h.unregister:
			if _, exist := h.clients[client.ID]; exist {
				if _, ok := h.clients[client.ID][client]; ok {
					delete(h.clients[client.ID], client)
					close(client.send)
				}
			}
		case message := <-h.broadcast:
			if h.isPrivate {
				msg := new(Message)
				json.Unmarshal(message, msg)
				b, _ := json.Marshal(msg.Data)
				if _, exist := h.clients[msg.ID]; exist {
					for client := range h.clients[msg.ID] {
						select {
						case client.send <- b:
						default:
							close(client.send)
							delete(h.clients[client.ID], client)
						}
					}
				}
			} else {
				for _, clients := range h.clients {
					for client := range clients {
						select {
						case client.send <- message:
						default:
							close(client.send)
							delete(h.clients[client.ID], client)
						}
					}
				}
			}
			
		}
	}
}