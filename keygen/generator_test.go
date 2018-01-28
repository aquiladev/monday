package keygen

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate1(t *testing.T) {
	p := generate(big.NewInt(1))

	assert.Equal(t, "5HpHagT65TZzG1PH3CSu63k8DbpvD8s5ip4nEB3kEsreAnchuDf", p.PrivKey)
	assert.Equal(t, "1EHNa6Q4Jz2uvNExL497mE43ikXhwF6kZm", p.PubKey)
	assert.Equal(t, "1BgGZ9tcN4rm9KBzDn7KprQz87SZ26SAMH", p.CompressedPubKey)
}

func TestGenerate128(t *testing.T) {
	//128 = 0x80
	p := generate(big.NewInt(128))

	assert.Equal(t, "5HpHagT65TZzG1PH3CSu63k8DbpvD8s5ip4nEB3kEsreR42AY81", p.PrivKey)
	assert.Equal(t, "1EoXPE6MzT4EnHvk2Ldj64M2ks2EAcZyH4", p.PubKey)
	assert.Equal(t, "1CAE6ej7VyAhgTtL1AYKTEByRJaCZKg8XM", p.CompressedPubKey)
}

func BenchmarkGenerate(b *testing.B) {
	for i := 1; i <= b.N; i++ {
		generate(big.NewInt(int64(i)))
	}
}
