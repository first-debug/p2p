package pb

import (
	"github.com/first-debug/p2p/internal/domain"
	"github.com/google/uuid"
)

func PbPeerToDomain(p *Peer) domain.Peer {
	return domain.Peer{
		ID:    uuid.UUID(p.ID.Value),
		Port:  int(p.Port),
		Files: p.Files,
	}
}

func DomainToPbPeer(p *domain.Peer) *Peer {
	// [uuid.UUID.MarshalBinary] never return not nil error
	return &Peer{
		ID:    ToPbUUID(p.ID),
		Port:  int32(p.Port),
		Files: p.Files,
	}
}

func ToPbUUID(id uuid.UUID) *UUID {
	bytes, _ := id.MarshalBinary()
	return &UUID{Value: bytes}
}
