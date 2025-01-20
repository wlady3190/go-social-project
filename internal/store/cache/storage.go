package cache

import (
	"github.com/go-redis/redis/v8"
	"github.com/wlady3190/go-social/internal/store"
	"golang.org/x/net/context"
)

type Storage struct {
	Users interface {
		Get(context.Context, int64) (*store.User, error)
		Set(context.Context, *store.User) error
		
		Delete(context.Context, int64)
	}
}

func NewRedisStorage(rdb *redis.Client) Storage {
	return Storage{
		Users: &UsersStore{
			rdb:rdb,
		},
	}
} //! DE aqui al main para activarlo
