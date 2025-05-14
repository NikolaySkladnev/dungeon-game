package domain

import "testing"

func TestNewPlayer(t *testing.T) {
	player := NewPlayer(42, 3)

	if player.ID != 42 {
		t.Fatalf("player ID = %d, want 42", player.ID)
	}
	if player.Health != MaxHealth {
		t.Fatalf("player health = %d, want %d", player.Health, MaxHealth)
	}
	if len(player.Floors) != 4 {
		t.Fatalf("len(player.Floors) = %d, want 4", len(player.Floors))
	}
	if player.Registered || player.InDungeon || player.Ended || player.Disqualified {
		t.Fatalf("new player has unexpected state: %+v", player)
	}
}
