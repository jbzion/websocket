package websocket

import (
	"github.com/go-redis/redis/v7"
	"github.com/gin-gonic/gin"
)

// Channels Manage a Hub and send and receive messages
type Channels struct {
	rdb				*redis.Client
	hub         	*Hub
	pubsub      	*redis.PubSub
	ChannelNames 	[]string
}

func (channels *Channels) run() {
	ch := channels.pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			channels.hub.broadcast <- []byte(msg.Payload)
		}
	}
}

// Received Returns a chan []byte that receives all messages 
// sent by the websocket connection
func (channels *Channels) Received() chan []byte {
	return channels.hub.received
}

// Middleware Returns a Gin HandleFunc that upgrader requests 
// to websocket requests 
func (channels *Channels) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("ws_channels", channels)
		conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		var id string
		value, ok := c.Get("ws_client_id")
		if ok {
			id = value.(string)
		}
		client := &Client{
			ID:   id,
			hub:  channels.hub,
			conn: conn,
			send: make(chan []byte, 256),
		}
		channels.hub.register <- client
		c.Next()
	}
}