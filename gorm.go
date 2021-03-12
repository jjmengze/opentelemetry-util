package opentelemetry

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/jjmengze/otgorm"
)

// RegisterGORMCallback registers the necessary callbacks in Gorm's hook system for instrumentation
func RegisterGORMCallback(db *gorm.DB) {
	otgorm.RegisterCallbacks(
		db,
		otgorm.Query(true),
		otgorm.Table(true),
		otgorm.AllowRoot(true),
	)
}

// GORMWithContext sets the current context in the db instance for instrumentation.
func GORMWithContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	return otgorm.WithContext(ctx, db)
}

type RepoInterface interface {
	ReadDB(ctx context.Context) *gorm.DB
	WriteDB(ctx context.Context) *gorm.DB
}

type Repo struct {
	_readDB  *gorm.DB
	_writeDB *gorm.DB
}

func NewRepo(readDB, writeDB *gorm.DB) *Repo {
	return &Repo{
		_readDB:  readDB,
		_writeDB: writeDB,
	}
}

func (repo *Repo) ReadDB(ctx context.Context) *gorm.DB {
	return GORMWithContext(ctx, repo._readDB)
}

func (repo *Repo) WriteDB(ctx context.Context) *gorm.DB {
	return GORMWithContext(ctx, repo._writeDB)
}
