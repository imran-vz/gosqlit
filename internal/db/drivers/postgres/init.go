package postgres

import "github.com/imran-vz/gosqlit/internal/db"

func init() {
	db.RegisterDriver("postgres", &Driver{})
}
