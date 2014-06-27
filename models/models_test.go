package models

import "testing"

func TestIsBoxAccessible(t *testing.T) {
	db := prepare(t)
	if a, err := IsBoxAccessible(db, boxes[0].Id, 2); err == nil || a {
		t.Error("should not be accessible", err)
	}

	if a, err := IsBoxAccessible(db, boxes[0].Id, 1); err != nil || !a {
		t.Error("should be accessible", err)
	}
}

func TestFetchBoxAndKeyPair(t *testing.T) {
	var b *Box
	var k *KeyPair
	var err error
	db := prepare(t)

	if _, _, err := FetchBoxAndKeyPair(db, boxes[0].Id, 2); err == nil {
		t.Error("should fail when fetch invalid box")
	}

	b, k, err = FetchBoxAndKeyPair(db, boxes[2].Id, boxes[2].OwnerId)
	if err == nil {
		t.Error("should miss keypair", err, b, k)
	}

	b, k, err = FetchBoxAndKeyPair(db, boxes[0].Id, 1)
	if err != nil || b.Id != boxes[0].Id || k.Id != keyPairs[0].Id {
		t.Error("should fetch box and keypair", err, b, k)
	}

	_, _, err = FetchBoxAndKeyPair(db, boxes[3].Id, 1)
	if err == nil {
		t.Error("should not fetch a stopped box")
	}
}
