package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
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

func (h *header) String() string {
	buf := bytes.NewBuffer(make([]byte, 0, 128))
	if err := binary.Write(buf, binary.LittleEndian, h); err != nil {
		panic(err)
	}
	b := buf.Bytes()
	b = append(b, 0x80)                        // 区切り
	b = append(b, make([]byte, 128-len(b))...) // 0埋め
	b[127] = 76 - 64                           // 末尾に0埋めされるbyte列の長さを追加

	var s strings.Builder
	for i, v := range b {
		switch {
		case i%32 == 0 && i != 0:
			s.WriteString("\n")
		case i%4 == 0 && i != 0:
			s.WriteString(" ")
		}
		if 72 <= i && i < 76 {
			s.WriteString("xx")
			continue
		}

		s.WriteString(fmt.Sprintf("%02x", v))
	}

	return s.String()
}

func main() {
	hash, h := fetchBlock(1)
	calculatedHash := h.calculateHash()
	fmt.Println(h.String())
	fmt.Println("expected hash  :", hash)
	fmt.Printf("calculated hash: %x\n", calculatedHash)

	if hash == hex.EncodeToString(calculatedHash) {
		fmt.Println("ハッシュの計算に成功")
	} else {
		panic("hash does not match")
	}

	if t := target(h.bits); bytes.Compare(calculatedHash, t) <= 0 {
		fmt.Printf("%x >= %x\nhash is less than target\n", t, calculatedHash)
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

func target(bits uint32) []byte {
	index := bits >> 24
	coefficient := bits & 0xffffff

	c := big.NewInt(int64(coefficient))
	return c.Lsh(c, uint(8*(index-3))).Bytes()
}
