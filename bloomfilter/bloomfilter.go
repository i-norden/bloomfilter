// Package bbloomfilter implements a bloomfilter based on an m-bit bit array, k hashfilters and configuration.
//
// It creates a bloomfilter based on bitset and an injected hasher, along with the
// following operations: add an element to the bloomfilter, check the existence of an element
// in the bloomfilter, the union of two bloomfilters, along with the serialization and
// deserialization of a bloomfilter: http://llimllib.github.io/bloomfilter-tutorial/
package bbloomfilter

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/krakendio/bloomfilter/v2"
	"github.com/tmthrgd/go-bitset"
)

// Bloomfilter basic type
type Bloomfilter struct {
	bs  bitset.Bitset
	m   uint
	k   uint
	h   []bloomfilter.Hash
	cfg bloomfilter.Config
}

// New creates a new bloomfilter from a given config
func New(cfg bloomfilter.Config) *Bloomfilter {
	m := bloomfilter.M(cfg.N, cfg.P)
	k := bloomfilter.K(m, cfg.N)
	return &Bloomfilter{
		m:   m,
		k:   k,
		h:   bloomfilter.HashFactoryNames[cfg.HashName](k),
		bs:  bitset.New(m),
		cfg: cfg,
	}
}

// Add an element to bloomfilter
func (b Bloomfilter) Add(elem []byte) {
	for _, h := range b.h {
		_, hs := h(elem)
		for _, x := range hs {
			b.bs.Set(x % b.m)
		}
	}
}

// Check if an element is in the bloomfilter
func (b Bloomfilter) Check(elem []byte) bool {
	for _, h := range b.h {
		_, hs := h(elem)
		for _, x := range hs {
			if !b.bs.IsSet(x % b.m) {
				return false
			}
		}
	}
	return true
}

func (b Bloomfilter) CheckWithReturn(elem []byte) ([]byte, bool) {
	res, phs := b.h[0](elem)
	for _, x := range phs {
		y := x % b.m
		if !b.bs.IsSet(y) {
			return res, false
		}
	}
	for _, h := range b.h[1:] {
		_, hs := h(elem)
		for _, x := range hs {
			y := x % b.m
			if !b.bs.IsSet(y) {
				return res, false
			}
		}
	}
	return res, true
}

// AddOrEject adds an element to the Bloomfilter if and only if its bits are not already set
// It returns the primary hash for the element and whether the element was added
func (b Bloomfilter) AddOrEject(elem []byte) ([]byte, bool) {
	set := false
	// this is the "primary hash", return it
	res, phs := b.h[0](elem)
	for _, x := range phs {
		y := x % b.m
		if !b.bs.IsSet(y) {
			set = true
			b.bs.Set(y)
		}
	}
	for _, h := range b.h[1:] {
		_, hs := h(elem)
		for _, x := range hs {
			y := x % b.m
			if !b.bs.IsSet(y) {
				// if it is not set, go ahead and set
				set = true
				b.bs.Set(y)
			}
		}
	}
	return res, set
}

// Union of two bloomfilters
func (b *Bloomfilter) Union(that interface{}) (float64, error) {
	other, ok := that.(*Bloomfilter)
	if !ok {
		return b.Capacity(), bloomfilter.ErrImpossibleToTreat
	}

	if b.m != other.m {
		return b.Capacity(), fmt.Errorf("m1(%d) != m2(%d)", b.m, other.m)
	}

	if b.k != other.k {
		return b.Capacity(), fmt.Errorf("k1(%d) != k2(%d)", b.k, other.k)
	}

	if b.cfg.HashName != other.cfg.HashName {
		return b.Capacity(), fmt.Errorf("different hashers: %s is not %s", other.cfg.HashName, b.cfg.HashName)
	}

	b.bs.Union(b.bs, other.bs)

	return b.Capacity(), nil
}

// SerializibleBloomfilter used when (de)serializing a bloomfilter
type SerializibleBloomfilter struct {
	BS       bitset.Bitset
	M        uint
	K        uint
	HashName string
	Cfg      bloomfilter.Config
}

// MarshalBinary serializes a bloomfilter
func (b *Bloomfilter) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(&SerializibleBloomfilter{
		BS:       b.bs,
		M:        b.m,
		K:        b.k,
		HashName: b.cfg.HashName,
		Cfg:      b.cfg,
	})
	//zip buf.Bytes

	return buf.Bytes(), err
}

// UnmarshalBinary deserializes a bloomfilter
func (b *Bloomfilter) UnmarshalBinary(data []byte) error {
	//unzip data
	buf := bytes.NewBuffer(data)
	target := SerializibleBloomfilter{}

	if err := gob.NewDecoder(buf).Decode(&target); err != nil {
		return err
	}
	*b = Bloomfilter{
		bs:  target.BS,
		m:   target.M,
		k:   target.K,
		h:   bloomfilter.HashFactoryNames[target.HashName](target.K),
		cfg: target.Cfg,
	}

	return nil
}

// Capacity returns the fill degree of the bloomfilter
func (b *Bloomfilter) Capacity() float64 {
	return float64(b.bs.Count()) / float64(b.m)
}

func (b *Bloomfilter) hashFactoryNameK(hashName string) []bloomfilter.Hash {
	return bloomfilter.HashFactoryNames[hashName](b.k)
}
