package main

import (
	"context"
	"fmt"
	"time"

	"github.com/William9923/redis-lib/pkg/cache"
)

type Dummy struct {
	UserID   int64  `json:"user_id" redis:"user_id"`
	Username string `json:"username" redis:"username"`
	Password string `json:"password" redis:"password"`
}

func main() {
	ctx := context.Background()
	rdb := cache.Connect(cache.Config{
		Address:  "localhost:6379",
		Password: "",
		DB:       0,
	})

	// userKey := "example:user"
	user := Dummy{
		UserID:   1,
		Username: "test-user",
		Password: "test-password",
	}

	// 1. Simulate save session
	sessionKey := "session:mock-key-session-1"
	reply := rdb.SetStructWithExpire(ctx, sessionKey, user, 30*time.Minute)
	if err := reply.Err(); err != nil {
		fmt.Println(err)
	}

	// 2. Simulate fetching session data
	var fetchedUser Dummy
	err := rdb.GetStruct(ctx, sessionKey, &fetchedUser)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Fetch from redis result...")
	fmt.Println(fetchedUser)
}
