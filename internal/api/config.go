package api

import (
	"sync/atomic"

	"github.com/inmydream21/chirpy/internal/database"
)

type APIConfig struct {
	FileserverHits atomic.Int64
	DB             *database.Queries
	Platform       string
	JWTSecret      string
	PolkaKey       string
}
