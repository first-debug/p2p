package pb

import (
	"net"

	"github.com/first-debug/p2p/internal/domain"
	"github.com/google/uuid"
)

func PbPeerToDomain(p *Peer) domain.Peer {
	return domain.Peer{
		ID:         uuid.UUID(p.ID.Value),
		IP:         net.ParseIP(p.IP),
		IsPublicIP: p.IsPublicIp,
		Port:       int(p.Port),
		Files:      p.Files,
	}
}

func DomainToPbPeer(p *domain.Peer) *Peer {
	return &Peer{
		ID:         ToPbUUID(p.ID),
		IP:         p.IP.String(),
		IsPublicIp: p.IsPublicIP,
		Port:       int32(p.Port),
		Files:      p.Files,
	}
}

func ToPbUUID(id uuid.UUID) *UUID {
	// [uuid.UUID.MarshalBinary] never return not nil error
	bytes, _ := id.MarshalBinary()
	return &UUID{Value: bytes}
}
