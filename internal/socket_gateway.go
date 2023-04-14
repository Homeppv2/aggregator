package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type SocketGateway struct {
	redis  *redis.Client
	prefix string
}

func NewSocketGateway(prefix string, redisURL string) *SocketGateway {
	return &SocketGateway{
		redis:  redis.NewClient(&redis.Options{Addr: redisURL}),
		prefix: prefix,
	}
}

func (s *SocketGateway) getKey(hw_key string) string {
	return fmt.Sprintf("%s:%s", s.prefix, hw_key)
}

func (s *SocketGateway) containsValue(ctx context.Context, key string, value string) (bool, error) {
	val, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return false, err
	} else if val == value {
		return true, nil
	}

	return false, nil
}

func (s *SocketGateway) SetConnected(ctx context.Context, hw_key string) error {
	key := s.getKey(hw_key)
	value := "connected"
	ok, err := s.containsValue(ctx, key, value)
	if err != nil {
		return err
	}
	if ok {
		return errors.New("socket already connected")
	}
	s.redis.Set(ctx, key, value, 0)

	return nil
}

func (s *SocketGateway) SetDisconnected(ctx context.Context, hw_key string) error {
	key := s.getKey(hw_key)
	value := "disconnected"
	ok, err := s.containsValue(ctx, key, value)
	if err != nil {
		return err
	}
	if ok {
		return errors.New("socket already disconnected")
	}
	s.redis.Set(ctx, key, value, 0)

	return nil
}

func (s *SocketGateway) IsConnected(ctx context.Context, hw_key string) (bool, error) {
	key := s.getKey(hw_key)
	value := "connected"
	ok, err := s.containsValue(ctx, key, value)
	if err != nil {
		return false, err
	}

	return ok, nil
}
