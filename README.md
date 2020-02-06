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
publicChannel := engine.PublicChannel("foo")	//Subscribe "foo" Redis pub/sub
r.GET("/ws/info", publicChannel.Middleware(), handle1)
```

handle1
```go
// Check authorization and rejection here
func handle1(c *gin.Context) {
	value, _ := c.Get("ws_channels")
	channels := value.(*websocket.Channels)

	received := channels.Received()
	for {
		select {
		case <-received:
		}
	}
}
```

## Private Channels

Router
```go
r := gin.Default()
privateChannel := engine.PrivateChannel("foo")
r.GET("/ws/order", handle1, privateChannel.Middleware(), handle2)
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
	value, _ := c.Get("ws_channels")
	channels := value.(*websocket.Channels)

	received := channels.Received()
	for {
		select {
		case <- received:
		}
	}
}
```

publish
```go
message := struct{
	ID	int64	`json:"id,string"`
	Data	interface{}	`json:"data"`
}{
	ID: 123,
	Data: "hello",	// ID只用來分配訊息，最終瀏覽器只會接收到Data
}
b, _ := json.Marshal(message)
rdb.Publish("foo", b)
```
