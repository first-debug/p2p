package sessionstorage

import (
	"context"

	"main/internal/session"

	"github.com/google/uuid"
)

type SessionStorage interface {
	Add(session.Session) error
	GetAll() ([]session.Session, error)
	// GetPage(start, stop int) ([]session.Session, error)
	GetByID(uuid.UUID) (session.Session, error)
	CloseByID(ctx context.Context, id uuid.UUID) error
	CloseAllByType(ctx context.Context, incoming bool) error
}
