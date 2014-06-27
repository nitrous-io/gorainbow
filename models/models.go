package models

import (
	"database/sql"
	"fmt"
)

type (
	Box struct {
		Id      int
		OwnerId int
		Host    string
		Port    int
		State   string
	}

	KeyPair struct {
		Id                    int
		BoxId                 int
		PublicKey             string
		EncryptedPrivateKey   string
		EncryptedPrivateKeyIv string
	}
)

const (
	accessibleQuery = "SELECT 1 FROM box_users WHERE box_id = $1 AND user_id = $2"
	boxKeyPairQuery = `SELECT
	b.id, b.owner_id, b.host, b.port,
	k.id, k.public_key, k.encrypted_private_key, k.encrypted_private_key_iv
	FROM boxes AS b INNER JOIN key_pairs AS k ON b.id = k.box_id
	WHERE b.id = $1
	AND b.state = 'running'
	AND EXISTS (SELECT 1 FROM box_users WHERE box_id = $2 AND user_id = $3)
	`
)

func IsBoxAccessible(db *sql.DB, boxId int, userId int) (bool, error) {
	var numUser int
	if err := db.QueryRow(accessibleQuery, boxId, userId).Scan(&numUser); err != nil {
		return false, err
	}
	return (numUser > 0), nil
}

func FetchBoxAndKeyPair(db *sql.DB, boxId int, userId int) (*Box, *KeyPair, error) {
	box := new(Box)
	keypair := new(KeyPair)

	err := db.QueryRow(boxKeyPairQuery, boxId, boxId, userId).Scan(
		&box.Id, &box.OwnerId, &box.Host, &box.Port,
		&keypair.Id, &keypair.PublicKey, &keypair.EncryptedPrivateKey, &keypair.EncryptedPrivateKeyIv,
	)
	if err != nil {
		return nil, nil, err
	}
	if box.Id == 0 {
		return nil, nil, fmt.Errorf("Box %v is not accessible for user %v", boxId, userId)
	}
	if keypair.Id == 0 {
		return nil, nil, fmt.Errorf("Box %v doesnt have keypair", boxId)
	}
	return box, keypair, nil
}
