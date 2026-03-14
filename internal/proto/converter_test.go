package pb

import (
	"testing"

	"github.com/first-debug/p2p/internal/domain"
	"github.com/google/uuid"
)

func TestPbPeerToDomain(t *testing.T) {
	id := uuid.New()

	pbPeer := &Peer{
		ID:   ToPbUUID(id),
		Port: 8080,
		Files: map[string]string{
			"file1.txt": "content1",
		},
	}

	peer := PbPeerToDomain(pbPeer)

	if peer.ID != id {
		t.Errorf("expected ID %v, got %v", id, peer.ID)
	}
	if peer.Port != 8080 {
		t.Errorf("expected Port 8080, got %d", peer.Port)
	}
	if len(peer.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(peer.Files))
	}
	if peer.Files["file1.txt"] != "content1" {
		t.Errorf("expected file content 'content1', got %s", peer.Files["file1.txt"])
	}
	if peer.IP != nil {
		t.Errorf("expected nil IP, got %v", peer.IP)
	}
}

func TestPbPeerToDomain_NilInput(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil input")
		}
	}()

	PbPeerToDomain(nil)
}

func TestDomainToPbPeer(t *testing.T) {
	id := uuid.New()

	domainPeer := &domain.Peer{
		ID:   id,
		Name: "test-peer",
		Port: 9090,
		Files: map[string]string{
			"doc.pdf": "binary",
		},
	}

	pbPeer := DomainToPbPeer(domainPeer)

	if pbPeer.ID == nil {
		t.Fatal("expected non-nil ID")
	}
	if pbPeer.Port != 9090 {
		t.Errorf("expected Port 9090, got %d", pbPeer.Port)
	}
	if len(pbPeer.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(pbPeer.Files))
	}
	if pbPeer.Files["doc.pdf"] != "binary" {
		t.Errorf("expected file content 'binary', got %s", pbPeer.Files["doc.pdf"])
	}
}

func TestDomainToPbPeer_NilInput(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil input")
		}
	}()

	DomainToPbPeer(nil)
}

func TestToPbUUID(t *testing.T) {
	id := uuid.New()

	pbUUID := ToPbUUID(id)

	if pbUUID == nil {
		t.Fatal("expected non-nil UUID")
	}

	recovered, err := uuid.FromBytes(pbUUID.Value)
	if err != nil {
		t.Fatalf("failed to recover UUID: %v", err)
	}
	if recovered != id {
		t.Errorf("expected UUID %v, got %v", id, recovered)
	}
}

func TestToPbUUID_NilUUID(t *testing.T) {
	pbUUID := ToPbUUID(uuid.Nil)

	if pbUUID == nil {
		t.Fatal("expected non-nil UUID")
	}

	recovered, err := uuid.FromBytes(pbUUID.Value)
	if err != nil {
		t.Fatalf("failed to recover UUID: %v", err)
	}
	if recovered != uuid.Nil {
		t.Errorf("expected nil UUID, got %v", recovered)
	}
}

func TestRoundTrip_UUID(t *testing.T) {
	original := uuid.New()

	pbUUID := ToPbUUID(original)
	bytes, err := uuid.FromBytes(pbUUID.Value)
	if err != nil {
		t.Fatalf("failed to convert back: %v", err)
	}

	if bytes != original {
		t.Errorf("round-trip failed: expected %v, got %v", original, bytes)
	}
}
