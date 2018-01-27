package keygen

import (
	"math/big"
	"testing"
)

func TestGenerate1(t *testing.T) {
	p := generate(big.NewInt(1))

	if p.PrivKey != "5HpHagT65TZzG1PH3CSu63k8DbpvD8s5ip4nEB3kEsreAnchuDf" ||
		p.PubKey != "1EHNa6Q4Jz2uvNExL497mE43ikXhwF6kZm" ||
		p.CompressedPubKey != "1BgGZ9tcN4rm9KBzDn7KprQz87SZ26SAMH" {
		t.Errorf("Wrong key")
	}
}

func TestGenerate128(t *testing.T) {
	//128 = 0x80
	p := generate(big.NewInt(128))

	if p.PrivKey != "5HpHagT65TZzG1PH3CSu63k8DbpvD8s5ip4nEB3kEsrek4AiGXU" ||
		p.PubKey != "1KsRadYLS39Wo7R6AJwZH6NDeX6w2V5pGP" ||
		p.CompressedPubKey != "1KNxuNtu6TR9fjLUir3WQbD54qJy5v6Ybe" {
		t.Errorf("Wrong key")
	}
}

func BenchmarkGenerate(b *testing.B) {
	for i := 1; i <= b.N; i++ {
		generate(big.NewInt(int64(i)))
	}
}
