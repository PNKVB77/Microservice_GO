package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mcrsrv_go/model"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func orderIdKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		fmt.Errorf("failed to encode order:%w", err)
	}

	key := orderIdKey(order.OrderId)

	txn := r.Client.TxPipeline()

	res := txn.SetNX(ctx, key, string(data), 0)

	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add to orders set:%w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec to redis:%w", err)
	}

	return nil
}

func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	key := orderIdKey(id)

	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Order{}, model.ErrNotExists
	} else if err != nil {
		return model.Order{}, fmt.Errorf("get an order:%w", err)
	}

	var order model.Order

	err = json.Unmarshal([]byte(value), &order)
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to decode order:%w", err)
	}

	return order, nil
}

func (r *RedisRepo) FindAll(ctx context.Context, page model.FindAllPage) (model.FindResult, error) {
	res := r.Client.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))
	keys, cursor, err := res.Result()
	if err != nil {
		return model.FindResult{}, fmt.Errorf("failed to find all order ids:%w", err)
	}
	if len(keys) == 0 {
		return model.FindResult{
			Orders: []model.Order{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return model.FindResult{}, fmt.Errorf("failed to find orders:%w", err)
	}

	orders := make([]model.Order, len(xs))

	for i, x := range xs {
		x, ok := x.(string)
		if !ok {
			continue
		}

		var order model.Order

		err = json.Unmarshal([]byte(x), &order)
		if err != nil {
			return model.FindResult{}, fmt.Errorf("failed to decode order:%w", err)
		}

		orders[i] = order
	}

	return model.FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}

func (r *RedisRepo) Delete(ctx context.Context, id uint64) error {
	key := orderIdKey(id)
	txn := r.Client.TxPipeline()

	err := r.Client.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return model.ErrNotExists
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("get an order:%w", err)
	}

	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to remove from order set:%w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec to redis:%w", err)
	}

	return nil
}

func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		fmt.Errorf("failed to encode order:%w", err)
	}
	key := orderIdKey(order.OrderId)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return model.ErrNotExists
	} else if err != nil {
		return fmt.Errorf("get an order:%w", err)
	}
	return nil
}
