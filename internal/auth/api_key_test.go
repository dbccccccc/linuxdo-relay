package auth

import "testing"

func TestGenerateUserAPIKey(t *testing.T) {
	plain, hashVal, err := GenerateUserAPIKey()
	if err != nil {
		t.Fatalf("generate api key failed: %v", err)
	}
	if len(plain) != len("sk-")+32 {
		t.Fatalf("unexpected key length: %d", len(plain))
	}
	if plain[:3] != "sk-" {
		t.Fatalf("key should start with sk- prefix")
	}
	if hashVal != HashAPIKey(plain) {
		t.Fatalf("hash mismatch")
	}
}
