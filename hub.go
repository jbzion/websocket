package websocket

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients 	map[*Client]struct{}
	// Broadcast to client
	broadcast 	chan []byte
	// Register requests from the clients.
	register 	chan *Client
	// Unregister requests from clients.
	unregister 	chan *Client
	//Message received from clients
	Received	chan []byte
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
			go client.writePump()
			go client.readPump()
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}