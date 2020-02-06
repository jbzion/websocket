package websocket

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"
	"net/http"
	"fmt"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Engine struct {
	rdb *redis.Client
}

// New Returns an Engine using Redis
func New(rdb *redis.Client) (*Engine, error) {
	_, err := rdb.Ping().Result()
	if err != nil {
		return nil, err
	}
	engine := &Engine{
		rdb: rdb,
	}
	return engine, nil
}

// Public Generate a Channels that subscribes to channelNames
// and return the Gin handlerFunc
func (engine *Engine) Public(channelNames ...string) gin.HandlerFunc {
	var pubsub *redis.PubSub
	if len(channelNames) != 0 {
		pubsub = engine.rdb.Subscribe(channelNames...)
		if _, err := pubsub.Receive(); err != nil {
			pubsub = nil
		}
	}
	channels := &Channels{
		rdb: engine.rdb,
		hub: &Hub{
			clients:    make(map[*Client]struct{}),
			broadcast:  make(chan []byte),
			register:   make(chan *Client),
			unregister: make(chan *Client),
			received:   make(chan []byte),
		},
		pubsub:       pubsub,
		ChannelNames: channelNames,
	}
	go channels.hub.run()
	if pubsub != nil {
		go channels.run()
	}
	return channels.middleware()
}

func (engine *Engine) Private() gin.HandlerFunc {
	return func(c *gin.Context) {
		channels := &Channels{
			rdb: engine.rdb,
			hub: &Hub{
				clients:    make(map[*Client]struct{}),
				broadcast:  make(chan []byte),
				register:   make(chan *Client),
				unregister: make(chan *Client),
				received:   make(chan []byte),
			},
		}
		go channels.hub.run()
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
		fmt.Println("pubsub close")
	}
}
