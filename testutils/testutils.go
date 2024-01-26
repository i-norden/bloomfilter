// Package testutils contains utils for the tests.
package testutils

import (
	"bytes"
	"testing"

	"golang.org/x/crypto/sha3"

	"github.com/krakendio/bloomfilter/v2"
)

var (
	TestCfg = bloomfilter.Config{
		N:        100,
		P:        0.001,
		HashName: bloomfilter.HASHER_OPTIMAL,
	}

	TestCfg2 = bloomfilter.Config{
		N:        100,
		P:        0.00001,
		HashName: bloomfilter.HASHER_OPTIMAL,
	}

	TestCfg3 = bloomfilter.Config{
		N:        100,
		P:        0.001,
		HashName: bloomfilter.HASHER_DEFAULT,
	}

	TestCfg4 = bloomfilter.Config{
		N:        100,
		P:        0.001,
		HashName: bloomfilter.HASHER_SECURE,
	}
)

func CallSet(t *testing.T, set bloomfilter.Bloomfilter) {
	set.Add([]byte{1, 2, 3})
	if !set.Check([]byte{1, 2, 3}) {
		t.Error("failed check")
	}

	if set.Check([]byte{1, 2, 4}) {
		t.Error("unexpected check")
	}
}

func CallSetOrEject(t *testing.T, set bloomfilter.EjectingBloomFilter) {
	set.AddOrEject([]byte{1, 2, 3})
	if !set.Check([]byte{1, 2, 3}) {
		t.Error("failed check")
	}

	if set.Check([]byte{1, 2, 4}) {
		t.Error("unexpected check")
	}

	h1, s1 := set.AddOrEject([]byte{1, 2, 3})
	if s1 {
		t.Error("duplicate should have been ejected")
	}

	h := sha3.NewLegacyKeccak256()
	h.Reset()
	h.Write([]byte{1, 2, 3})
	expectedHash := h.Sum(nil)

	if !bytes.Equal(h1, expectedHash) {
		t.Error("AddOrEject returned hash does not equal expected hash")
	}

	h2, s2 := set.CheckWithReturn([]byte{1, 2, 3})
	if !s2 {
		t.Error("expected CheckWithReturn to return true")
	}
	if !bytes.Equal(h2, expectedHash) {
		t.Error("returned hash does not equal expected hash")
	}

	h3, s3 := set.CheckWithReturn([]byte{1, 2, 4})
	if s3 {
		t.Error("unexpected set")
	}

	h.Reset()
	h.Write([]byte{1, 2, 4})
	expectedHash = h.Sum(nil)

	if !bytes.Equal(h3, expectedHash) {
		t.Error("CheckWithReturn returned hash does not equal expected hash")
	}
}

func CallSetUnion(t *testing.T, set1, set2 bloomfilter.Bloomfilter) {
	elem := []byte{1, 2, 3}
	set1.Add(elem)
	if !set1.Check(elem) {
		t.Error("failed add set1 before union")
		return
	}

	if set2.Check(elem) {
		t.Error("unexpected check to union of set2")
		return
	}

	if _, err := set2.Union(set1); err != nil {
		t.Error("failed union set1 to set2", err.Error())
		return
	}

	if !set2.Check(elem) {
		t.Error("failed union check of set1 to set2")
		return
	}
}
