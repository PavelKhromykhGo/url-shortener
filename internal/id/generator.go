package id

import (
	"crypto/rand"
	"fmt"
)

type Generator interface {
	GenerateShortCode() (string, error)
}

type RandomGenerator struct {
	length       int
	allowedChars []byte
}

func NewRandomGenerator(length int) *RandomGenerator {
	allowed := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	return &RandomGenerator{
		length:       length,
		allowedChars: allowed,
	}
}

func (g *RandomGenerator) GenerateShortCode() (string, error) {
	b := make([]byte, g.length)

	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("crypto rand read failed %w", err)
	}

	for i := range b {
		b[i] = g.allowedChars[int(b[i])%len(g.allowedChars)]
	}
	return string(b), nil
}
