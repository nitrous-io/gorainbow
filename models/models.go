package models

import (
	"database/sql"
	"fmt"
)

type Box struct {
	Id      int
	OwnerId int
	Host    string
	Port    int
}

type KeyPair struct {
	Id                    int
	PublicKey             string
	EncryptedPrivateKey   string
	EncryptedPrivateKeyIv string
}

func FetchBox(db *sql.DB, boxId int) (*Box, error) {
	box := new(Box)
	if err := db.QueryRow(
		"SELECT id, owner_id, host, port FROM boxes WHERE id = $1",
		boxId).Scan(&box.Id, &box.OwnerId, &box.Host, &box.Port); err != nil {
		return nil, err
	}
	return box, nil
}

func FetchKeyPair(db *sql.DB, boxId int) (*KeyPair, error) {
	keypair := new(KeyPair)
	err := db.QueryRow(
		"SELECT id, public_key, encrypted_private_key, encrypted_private_key_iv FROM key_pairs WHERE box_id = $1",
		boxId).Scan(&keypair.Id, &keypair.PublicKey,
		&keypair.EncryptedPrivateKey, &keypair.EncryptedPrivateKeyIv)
	if err != nil {
		return nil, err
	}
	return keypair, nil
}

func IsBoxAccessible(db *sql.DB, boxId int, userId int) (bool, error) {
	var numUser int
	if err := db.QueryRow(
		"SELECT count(*) FROM box_users WHERE box_id = $1 AND user_id = $2",
		boxId, userId).Scan(&numUser); err != nil {
		return false, err
	}
	return (numUser > 0), nil
}

func FetchBoxAndKeyPair(db *sql.DB, boxId int, userId int) (*Box, *KeyPair, error) {
	accessible, err := IsBoxAccessible(db, boxId, userId)
	if err != nil {
		return nil, nil, err
	}
	if !accessible {
		err := fmt.Errorf("Box %v is not accessible for user %v", boxId, userId)
		return nil, nil, err
	}
	box, err := FetchBox(db, boxId)
	if err != nil {
		return nil, nil, err
	}
	keypair, err := FetchKeyPair(db, boxId)
	if err != nil {
		return nil, nil, err
	}
	return box, keypair, nil
}
