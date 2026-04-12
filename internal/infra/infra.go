// Package infra aggregates all infrastructure Wire provider sets and
// exposes application startup helpers (migration, cache, metrics).
package infra

import "github.com/google/wire"

// Provider is the single Wire provider set for all infrastructure dependencies.
// It supplies: *gorm.DB, *redis.Client (db), *Loggers (logger).
var Provider = wire.NewSet(dbSet, loggerSet)
