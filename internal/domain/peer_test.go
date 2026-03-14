package domain

import (
	"net"
	"testing"

	"github.com/google/uuid"
)

func TestPeer_Fields(t *testing.T) {
	id := uuid.New()
	testIP := net.ParseIP("192.168.1.1")

	peer := Peer{
		ID:   id,
		Name: "test-peer",
		IP:   testIP,
		Port: 8080,
		Files: map[string]string{
			"file1.txt": "content1",
			"file2.txt": "content2",
		},
	}

	if peer.ID != id {
		t.Errorf("expected ID %v, got %v", id, peer.ID)
	}
	if peer.Name != "test-peer" {
		t.Errorf("expected Name 'test-peer', got %s", peer.Name)
	}
	if !peer.IP.Equal(testIP) {
		t.Errorf("expected IP %v, got %v", testIP, peer.IP)
	}
	if peer.Port != 8080 {
		t.Errorf("expected Port 8080, got %d", peer.Port)
	}
	if len(peer.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(peer.Files))
	}
	if peer.Files["file1.txt"] != "content1" {
		t.Errorf("expected file1.txt content 'content1', got %s", peer.Files["file1.txt"])
	}
}

func TestPeer_EmptyFields(t *testing.T) {
	peer := Peer{}

	if peer.ID != uuid.Nil {
		t.Errorf("expected empty ID, got %v", peer.ID)
	}
	if peer.Name != "" {
		t.Errorf("expected empty Name, got %s", peer.Name)
	}
	if peer.IP != nil {
		t.Errorf("expected nil IP, got %v", peer.IP)
	}
	if peer.Port != 0 {
		t.Errorf("expected empty Port 0, got %d", peer.Port)
	}
	if peer.Files != nil {
		t.Errorf("expected nil Files map, got %v", peer.Files)
	}
}
