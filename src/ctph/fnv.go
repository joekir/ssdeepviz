package ctph

// Custom implementation of FNV32 based off the implementation used in SSDEEP:
// https://github.com/ssdeep-project/ssdeep/blob/master/fuzzy.c#L109
//
// Implementation copied almost entirely from stdlib hash/fnv
//
// The primary difference is the use of a non-standard FNV offset, 0x28021967.

const (
	offset = 0x28021967 // SSDEEP specific FNV offset value
	prime  = 16777619   // Standard FNV 32 bit prime
)

type Sum32 uint32

// Generate a new FNV32 with the custom SSDEEP offset.
func NewFNV() *Sum32 {
	var s Sum32 = offset
	return &s
}

// Reset the hash to the custom SSDEEP offset.
func (s *Sum32) Reset() { *s = offset }

//////////////////////////////////////////////////////////////////////////////
// Everything from this point on is identical to stdlib FNV32 implementations.
//////////////////////////////////////////////////////////////////////////////
func (s *Sum32) Write(data []byte) (int, error) {
	hash := *s
	for _, c := range data {
		hash *= prime
		hash ^= Sum32(c)
	}
	*s = hash
	return len(data), nil
}

func (s *Sum32) Size() int { return 4 }

func (s *Sum32) Sum(in []byte) []byte {
	v := uint32(*s)
	return append(in, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func (s *Sum32) Sum32() uint32 { return uint32(*s) }

func (s *Sum32) BlockSize() int { return 1 }
