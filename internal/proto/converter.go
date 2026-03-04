package pb

import "github.com/first-debug/p2p/internal/domain"

func PbPeerToDomain(p *Peer) domain.Peer {
	return domain.Peer{
		ID:    p.Id,
		Port:  int(p.Port),
		Files: p.Files,
	}
}

func DomainToPbPeer(p *domain.Peer) *Peer {
	return &Peer{
		Id:    p.ID,
		Port:  int32(p.Port),
		Files: p.Files,
	}
}
