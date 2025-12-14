package id

import (
	"crypto/rand"
	"fmt"
)

// Generator produces short codes for shortened URLs.
type Generator interface {
	GenerateShortCode() (string, error)
}

// RandomGenerator uses crypto/rand to build random short codes from an allowed alphabet.
type RandomGenerator struct {
	length       int
	allowedChars []byte
}

// NewRandomGenerator creates a generator that builds codes with the given length.
func NewRandomGenerator(length int) *RandomGenerator {
	allowed := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	return &RandomGenerator{
		length:       length,
		allowedChars: allowed,
	}
}

// GenerateShortCode produces a new random short code.
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
