package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/wlady3190/go-social/internal/store"
	"golang.org/x/net/context"
)

type UsersStore struct {
	rdb *redis.Client
}

const UserExpTime = time.Minute

func (s *UsersStore) Get(ctx context.Context, userID int64) (*store.User, error) {
	if s.rdb == nil {
		return nil, fmt.Errorf("redis client is not initialized")
	}
	cacheKey := fmt.Sprintf("user-%d", userID)
	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	var user store.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}
	return &user, nil
}

func (s *UsersStore) Set(ctx context.Context, user *store.User) error {
	//TTL
	cacheKey := fmt.Sprintf("user-%d", user.ID)
	json, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.rdb.SetEX(ctx, cacheKey, json, UserExpTime).Err()

}

//! Esto se aplica en el miidleware, en la obtención del Users.GetById

func (s *UsersStore) Delete(ctx context.Context, userID int64) {
	//! Eliminar info del cache
	cacheKey := fmt.Sprintf("user-%d", userID)
	s.rdb.Del(ctx, cacheKey)
}