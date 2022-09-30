//go:build go1.18
// +build go1.18

package surrealdb

import "context"

//Simple generic wrappers around db methods for convenience / code cleanliness
//todo: not really performant atm

func QueryAs[T any](ctx *context.Context, db *DB, sql string, vars any) (*T, error) {
	var container T
	err := db.Query(ctx, sql, vars).Unmarshal(&container)
	if err != nil {
		if err == ErrNoResult {
			return nil, nil
		}
		return nil, err
	}
	return &container, nil
}

func SelectAs[T any](ctx *context.Context, db *DB, what string) (*T, error) {
	var container T
	err := db.Select(ctx, what).Unmarshal(&container)
	if err != nil {
		if err == ErrNoResult {
			return nil, nil
		}
		return nil, err
	}
	return &container, nil
}

func CreateAs[T any](ctx *context.Context, db *DB, thing string, data any) (*T, error) {
	var container T
	err := db.Create(ctx, thing, data).Unmarshal(&container)
	if err != nil {
		if err == ErrNoResult {
			return nil, nil
		}
		return nil, err
	}
	return &container, nil
}

func UpdateAs[T any](ctx *context.Context, db *DB, what string, data any) (*T, error) {
	var container T
	err := db.Update(ctx, what, data).Unmarshal(&container)
	if err != nil {
		if err == ErrNoResult {
			return nil, nil
		}
		return nil, err
	}
	return &container, nil
}

func ChangeAs[T any](ctx *context.Context, db *DB, what string, data any) (*T, error) {
	var container T
	err := db.Change(ctx, what, data).Unmarshal(&container)
	if err != nil {
		if err == ErrNoResult {
			return nil, nil
		}
		return nil, err
	}
	return &container, nil
}

func ModifyAs[T any](ctx *context.Context, db *DB, what string, data []Patch) (*T, error) {
	var container T
	err := db.Modify(ctx, what, data).Unmarshal(&container)
	if err != nil {
		if err == ErrNoResult {
			return nil, nil
		}
		return nil, err
	}
	return &container, nil
}
