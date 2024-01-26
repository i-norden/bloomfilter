package bloomfilter

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"hash"
	"hash/crc64"
	"hash/fnv"

	"golang.org/x/crypto/sha3"
)

type Hash func([]byte) ([]byte, []uint)
type HashFactory func(uint) []Hash

const (
	HASHER_DEFAULT = "default"
	HASHER_OPTIMAL = "optimal"
	HASHER_SECURE  = "secure"
)

var (
	defaultHashers = []Hash{
		MD5,
		CRC64,
		SHA1,
		FNV64,
		FNV128,
	}

	secureHashers = append([]Hash{
		KECCAK256,
	}, defaultHashers...)

	HashFactoryNames = map[string]HashFactory{
		HASHER_DEFAULT: DefaultHashFactory,
		HASHER_OPTIMAL: OptimalHashFactory,
		HASHER_SECURE:  SecureHashFactory,
	}

	ErrImpossibleToTreat = fmt.Errorf("unable to union")

	KECCAK256 = HashWrapper(sha3.NewLegacyKeccak256())
	MD5       = HashWrapper(md5.New())
	SHA1      = HashWrapper(sha1.New())
	CRC64     = HashWrapper(crc64.New(crc64.MakeTable(crc64.ECMA)))
	FNV64     = HashWrapper(fnv.New64())
	FNV128    = HashWrapper(fnv.New128())
)

func DefaultHashFactory(k uint) []Hash {
	if k > uint(len(defaultHashers)) {
		k = uint(len(defaultHashers))
	}
	return defaultHashers[:k]
}

func SecureHashFactory(k uint) []Hash {
	if k > uint(len(secureHashers)) {
		k = uint(len(secureHashers))
	}
	return secureHashers[:k]
}

func OptimalHashFactory(k uint) []Hash {
	return []Hash{
		func(b []byte) ([]byte, []uint) {
			res, hs := FNV128(b)
			out := make([]uint, k)

			for i := range out {
				out[i] = hs[0] + uint(i)*hs[1]
			}
			return res, out
		},
	}
}

func HashWrapper(h hash.Hash) Hash {
	return func(elem []byte) ([]byte, []uint) {
		h.Reset()
		h.Write(elem)
		result := h.Sum(nil)
		out := make([]uint, len(result)/8)
		for i := 0; i < len(result)/8; i++ {
			out[i] = uint(binary.LittleEndian.Uint64(result[i*8 : (i+1)*8]))
		}
		return result, out
	}
}
