package websocket

import (
	"fmt"
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

// Subscribe to channelNames and run the channel
func (channels *Channels) Subscribe(channelNames ...string) error {
	if channels.pubsub != nil {
		return fmt.Errorf("Channels are active, subscriptions cannot be changed")
	}
	pubsub := channels.rdb.Subscribe(channelNames...)
	if _, err := pubsub.Receive(); err != nil {
		return err
	}
	channels.pubsub = pubsub
	channels.ChannelNames = channelNames
	go channels.run()
	return nil
}

// Publish Send message to all subscriptions on this channels
func (channels *Channels) Publish(msg string) {
	for _, v := range channels.ChannelNames {
		channels.rdb.Publish(v, msg)
	}
}

// PublishOther Send message to other Redis pubsub channel
func (channels *Channels) PublishOther(channelName string, msg string) {
	channels.rdb.Publish(channelName, msg)
}

// Received Returns a chan []byte that receives all messages 
// sent by the websocket connection
func (channels *Channels) Received() chan []byte {
	return channels.hub.received
}

// middleware Returns a Gin HandleFunc that upgrader requests 
// to websocket requests 
func (channels *Channels) middleware() gin.HandlerFunc {
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
		channels.pubsub.Close()
	}
}
