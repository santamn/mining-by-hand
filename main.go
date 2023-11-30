package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type header struct {
	version        uint32
	hashPrevBlock  [32]byte
	hashMerkleRoot [32]byte
	time           uint32
	bits           uint32
	nonce          uint32
}

func (h *header) calculateHash() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 76))
	if err := binary.Write(buf, binary.LittleEndian, h); err != nil {
		panic(err)
	}
	hash := doubleSHA256(buf.Bytes())
	return reverse(hash[:])
}

func main() {
	hash, h := fetchBlock(800000)
	calculatedHash := h.calculateHash()
	fmt.Println("expected hash  :", hash)
	fmt.Printf("calculated hash: %x\n", calculatedHash)

	if target(h.bits) >= binary.BigEndian.Uint64(calculatedHash) {
		fmt.Println("hash is less than target")
	} else {
		panic("hash is not less than target")
	}
}

func fetchBlock(height uint) (hash string, h *header) {
	type block struct {
		Hash       string `json:"hash"`
		Version    uint32 `json:"ver"`
		PrevBlock  string `json:"prev_block"`
		MerkleRoot string `json:"mrkl_root"`
		Time       uint32 `json:"time"`
		Bits       uint32 `json:"bits"`
		Nonce      uint32 `json:"nonce"`
	}

	resp, err := http.Get(fmt.Sprintf("https://blockchain.info/block-height/%d?format=json", height))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	m := make(map[string][]block)
	if err := json.Unmarshal(body, &m); err != nil {
		panic(err)
	}

	b := m["blocks"][0]
	return b.Hash, &header{
		version:        b.Version,
		hashPrevBlock:  decodeHexToLittleEndian(b.PrevBlock),
		hashMerkleRoot: decodeHexToLittleEndian(b.MerkleRoot),
		time:           b.Time,
		bits:           b.Bits,
		nonce:          b.Nonce,
	}
}

func doubleSHA256(b []byte) [32]byte {
	hash1 := sha256.Sum256(b)
	return sha256.Sum256(hash1[:])
}

func reverse[T any](s []T) []T {
	r := make([]T, len(s))
	for i := range s {
		r[len(s)-1-i] = s[i]
	}
	return r
}

func decodeHexToLittleEndian(s string) [32]byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	if len(b) != 32 {
		panic("invalid length")
	}

	var r [32]byte
	copy(r[:], reverse(b))
	return r
}

func target(bits uint32) uint64 {
	index := bits >> 24
	coefficient := bits & 0xffffff
	return uint64(coefficient) << (8 * (index - 3))
}
