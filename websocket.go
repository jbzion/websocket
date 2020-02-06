package websocket

import (
	"net/http"

	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"
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

// PublicChannel Generate a Channels that subscribes to channelNames
func (engine *Engine) PublicChannel(channelNames ...string) *Channels {
	return engine.newChannel(false, channelNames...)
}

func (engine *Engine) PrivateChannel(channelNames ...string) *Channels {
	return engine.newChannel(true, channelNames...)
}

func (engine *Engine) newChannel(isPrivate bool, channelNames ...string) *Channels {
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
			isPrivate:	isPrivate,
		},
		pubsub:       pubsub,
		ChannelNames: channelNames,
	}
	go channels.hub.run()
	if pubsub != nil {
		go channels.run()
	}
	return channels
}
