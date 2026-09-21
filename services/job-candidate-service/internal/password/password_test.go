package password

import "testing"

func TestHashAndVerify(t *testing.T) {
	plain := "supersecret123"
	hash, err := Hash(plain)
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}
	if hash == plain {
		t.Fatal("Hash() returned plain text")
	}
	if !Verify(plain, hash) {
		t.Fatal("Verify() should return true for correct password")
	}
	if Verify("wrongpassword", hash) {
		t.Fatal("Verify() should return false for wrong password")
	}
}

func TestHash_DifferentSalts(t *testing.T) {
	plain := "samepassword"
	hash1, _ := Hash(plain)
	hash2, _ := Hash(plain)
	if hash1 == hash2 {
		t.Fatal("Hash() should produce different hashes due to random salt")
	}
}
