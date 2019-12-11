# websocket
Convenient, using Redis Pub/sub, Websocket lib for decentralized systems, kubenetes

## Quick start

Import
```go
import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/eric11jhou/websocket"
)
```

Create Redis client
```go
rdb := redis.NewClient(&redis.Options{
  Addr:     "localhost:6379",
  Password: "", // no password set
  DB:       0,  // use default DB
})
```

Create websocket engine
```go
engine, _ := websocket.New(rdb)
```

## Public Channels

Router
```go
r := gin.Default()
r.GET("/ws/:user_id", handle1, engine.Public("foo"), handle2) //Subscribe "foo" Redis pub/sub
```

handle1
```go
// Check authorization and rejection here
func handle1(c *gin.Context) {
	id := c.Param("user_id")  // You can also get the id from the database
	c.Set("ws_client_id", id) // ws_client_id is the prefix for client sending messages like bar:message
	c.Next()
}
```

handle2
```go
// Handle received messages here if you need to
func handle2(c *gin.Context) {
	value, _ := c.Get("ws_channels")
	channels := value.(*websocket.Channels)
  
	received := channels.Received()
	for {
		select {
		case msg := <- received:  //msg like bar:message
			channels.Publish(string(msg)) //send msg to Redis pub/sub that foo in engine.Public("foo")
			channels.PublishOther(otherName, string(msg))  //send msg to Other Redis pub/sub
		}
	}
}
```

## Private Channels

Router
```go
r := gin.Default()
r.GET("/ws/:room_id/:user_id", handle1, engine.Private(), handle2)
```

handle1
```go
// Check authorization and rejection here
func handle1(c *gin.Context) {
	id := c.Param("user_id")  // You can also get the id from the database
	c.Set("ws_client_id", id) // ws_client_id is the prefix for client sending messages like bar:message
	c.Next()
}
```

handle2
```go
// Subscribe to Redis pub/sub here
func handle2(c *gin.Context) {
	room_id := c.Param("room_id") // You can also get the id from the database
  
	value, _ := c.Get("ws_channels")
	channels := value.(*websocket.Channels)
  
	channels.Subscribe(room_id) // Subscribe to Redis pub/sub
	received := channels.Received()
	for {
		select {
		case msg := <- received:
			channels.Publish(string(msg)) //send msg to Redis pub/sub that foo in engine.Public("foo")
			channels.PublishOther(otherName, string(msg))  //send msg to Other Redis pub/sub
		}
	}
}
```
