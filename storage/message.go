package storage

type KeyPair struct {
	PrivateKey string `json:"p"`
	PublicKey  string `json:"k"`
}

type Message struct {
	Keys []*KeyPair `json:"i"`
}
