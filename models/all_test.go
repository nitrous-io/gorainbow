package models

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

var dbStructure = []string{
	// box data structure
	`DROP TABLE IF EXISTS boxes;`,
	`CREATE TABLE boxes (
		id integer NOT NULL,
		owner_id integer NOT NULL,
		port integer NOT NULL,
		host character varying(255));`,
	`CREATE SEQUENCE boxes_id_seq START WITH 1 INCREMENT BY 1
		NO MINVALUE NO MAXVALUE CACHE 1;`,
	`ALTER SEQUENCE boxes_id_seq OWNED BY boxes.id;`,
	`ALTER TABLE ONLY boxes ALTER COLUMN id SET DEFAULT nextval('boxes_id_seq'::regclass);`,
	`ALTER TABLE ONLY boxes ADD CONSTRAINT boxes_pkey PRIMARY KEY (id);`,

	// box_users data structure
	`DROP TABLE IF EXISTS box_users;`,
	`CREATE TABLE box_users (
		id integer NOT NULL,
		box_id integer NOT NULL,
		user_id integer NOT NULL);`,
	`CREATE SEQUENCE box_users_id_seq START WITH 1 INCREMENT BY 1
		NO MINVALUE NO MAXVALUE CACHE 1;`,
	`ALTER SEQUENCE box_users_id_seq OWNED BY box_users.id;`,
	`ALTER TABLE ONLY box_users ALTER COLUMN id SET DEFAULT nextval('box_users_id_seq'::regclass);`,
	`ALTER TABLE ONLY box_users ADD CONSTRAINT box_users_pkey PRIMARY KEY (id);`,

	// key_pairs data structure
	`DROP TABLE IF EXISTS key_pairs;`,
	`CREATE TABLE key_pairs (
		id integer NOT NULL,
		box_id integer NOT NULL,
		public_key text NOT NULL,
		encrypted_private_key text NOT NULL,
		encrypted_private_key_iv character varying(255) NOT NULL);`,
	`CREATE SEQUENCE key_pairs_id_seq START WITH 1 INCREMENT BY 1
		NO MINVALUE NO MAXVALUE CACHE 1;`,
	`ALTER SEQUENCE key_pairs_id_seq OWNED BY key_pairs.id;`,
	`ALTER TABLE ONLY key_pairs ALTER COLUMN id SET DEFAULT nextval('key_pairs_id_seq'::regclass);`,
	`ALTER TABLE ONLY key_pairs ADD CONSTRAINT key_pairs_pkey PRIMARY KEY (id);`,
}

type boxUser struct {
	Id     int
	BoxId  int
	UserId int
}

var (
	boxes    []*Box
	boxUsers []*boxUser
	keyPairs []*KeyPair
)

// create database connection and prepare fixtures
func prepare(t *testing.T) *sql.DB {
	db, _ := sql.Open("postgres", "host=/var/run/postgresql dbname=testdb sslmode=disable")
	for _, s := range dbStructure {
		if _, err := db.Exec(s); err != nil {
			t.Fatal("failed to set db structure", err, s)
		}
	}

	boxes = []*Box{
		{OwnerId: 1, Port: 1001, Host: "apse1.nitrousbox.com"},
		{OwnerId: 1, Port: 1002, Host: "use1.nitrousbox.com"},
		{OwnerId: 2, Port: 1003, Host: "apse1.nitrousbox.com"},
	}
	for _, b := range boxes {
		err := db.QueryRow(
			"INSERT INTO boxes (owner_id, port, host) VALUES ($1, $2, $3) RETURNING ID",
			b.OwnerId, b.Port, b.Host).Scan(&b.Id)
		if err != nil {
			t.Fatal("failed to insert boxes", err)
		}
	}

	boxUsers = []*boxUser{
		{BoxId: boxes[0].Id, UserId: 1},
		{BoxId: boxes[1].Id, UserId: 1},
		{BoxId: boxes[2].Id, UserId: 2},
	}
	for _, b := range boxUsers {
		err := db.QueryRow(
			"INSERT INTO box_users (box_id, user_id) VALUES ($1, $2) RETURNING ID",
			b.BoxId, b.UserId).Scan(&b.Id)
		if err != nil {
			t.Fatal("failed to insert box_users", err)
		}
	}

	keyPairs = []*KeyPair{
		{BoxId: boxes[0].Id, PublicKey: "pubkey1", EncryptedPrivateKey: "priv1", EncryptedPrivateKeyIv: "iv1"},
		{BoxId: boxes[1].Id, PublicKey: "pubkey2", EncryptedPrivateKey: "priv2", EncryptedPrivateKeyIv: "iv2"},
	}
	for _, v := range keyPairs {
		err := db.QueryRow(
			"INSERT INTO key_pairs (box_id, public_key, encrypted_private_key, encrypted_private_key_iv) VALUES ($1, $2, $3, $4) RETURNING ID",
			v.BoxId, v.PublicKey, v.EncryptedPrivateKey,
			v.EncryptedPrivateKeyIv).Scan(&v.Id)
		if err != nil {
			t.Fatal("failed to insert key pairs", err)
		}
	}
	return db
}
