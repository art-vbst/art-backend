package postgres

import (
	"github.com/art-vbst/art-backend/internal/platform/db/store"
)

type Postgres struct {
	db *store.Store
}

func New(db *store.Store) *Postgres {
	return &Postgres{db: db}
}
