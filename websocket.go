package websocket

import (
	"net/http/
http"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"
)

type Engine struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Engine {
	engine := &Engine{
		rdb: rdb,
	}

	return engine
}

func (engine *Engine) Global(channelName string) (*Channel, error) {
	pubsub := engine.rdb.Subscribe(channelName)
	if _, err := pubsub.Receive(); err != nil {
		return nil, err
	}
	channel := &Channel{
		hub: &Hub{
			clients:    make(map[*Client]struct{}),
			broadcast:  make(chan []byte),
			register:   make(chan *Client),
			unregister: make(chan *Client),
			Received:   make(chan []byte),
		},
		pubsub:      pubsub,
		ChannelName: channelName,
		ChannelKey:  "",
	}
	go channel.run()
	return channel, nil
}

type Channel struct {
	hub         *Hub
	pubsub      *redis.PubSub
	ChannelName string
	ChannelKey  string
}

func (channel *Channel) run() {
	go channel.hub.run()
	ch := channel.pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			channel.hub.broadcast <- []byte(msg.Payload)
		}
	}
}

func (channel *Channel) Middleware() gin.HandlerFunc {
	var upGrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return func(c *gin.Context) {
		c.Keys["channel"] = channel
		conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		var id string
		value, ok := c.Keys["conn_id"]
		if ok {
			id = value.(string)
		}
		client := &Client{
			ID:   id,
			hub:  channel.hub,
			conn: conn,
			send: make(chan []byte, 256),
		}
		channel.hub.register <- client
		c.Next()
	}
}
