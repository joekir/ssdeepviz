package ctph

import (
	"errors"
	"fmt"
	"github.com/agnivade/levenshtein"
	"regexp"
	"strings"
)

const (
	B64_CHARS      string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	SS_LEN         uint32 = 64
	SS_SIG_PATTERN string = "^\\d+:[0-9a-zA-Z+\\/]+:[0-9a-zA-Z+\\/]+$"
	WINDOW_SIZE    uint32 = 7
	blockSizeMin   uint32 = 3
)

// Implementation based on https://github.com/ssdeep-project/ssdeep/blob/master/fuzzy.c#L383
func calcInitBlockSize(u uint32) uint32 {
	var bi uint32
	for (blockSizeMin<<bi)*SS_LEN < u {
		bi++
	}

	return blockSizeMin << bi
}

/*
* Compares 2 CTPH Signatures (s1 and s2) of the form:
* 24:O7XC9FZ2LBfaW3h+XdcDljuQJtNMMqF5DjQuwM0OHC:O7S9FZ2LwWEdcM6tNMjDEuwwHC
* <blocksize>:<sigpart1>:<sigpart2>
*
* The 2006 paper recommends using levenshtein distance to compare
*
* This function will return the distance as a positive integer
 */
func Compare(s1, s2 string) (int, error) {
	if matched, err := regexp.MatchString(SS_SIG_PATTERN, s1); err != nil || !matched {
		return -1, errors.New("invalid pattern in string 1")
	}

	if matched, err := regexp.MatchString(SS_SIG_PATTERN, s2); err != nil || !matched {
		return -1, errors.New("invalid pattern in string 2")
	}

	Sig1 := strings.Split(s1, ":")
	Sig2 := strings.Split(s2, ":")

	if Sig1[0] != Sig2[0] {
		return -1, errors.New("blocksize mismatch")
	}

	firstDiff := levenshtein.ComputeDistance(Sig1[1], Sig2[1])
	secondDiff := levenshtein.ComputeDistance(Sig1[2], Sig2[2])

	if secondDiff < firstDiff {
		return secondDiff, nil
	}

	return firstDiff, nil
}

/*
 * State of the fuzzy hash that can be manipulated by the
 * APIs attached to it
 */
type FuzzyHash struct {
	Bs         uint32      `json:"block_size"`
	Hash1      Sum32       `json:"-"`
	Hash2      Sum32       `json:"-"`
	Index      int         `json:"index"`
	InputLen   int         `json:"input_length"`
	IsTrigger1 bool        `json:"is_trigger1"`
	IsTrigger2 bool        `json:"is_trigger2"`
	Retry      bool        `json:"-"`
	Rh         RollingHash `json:"rolling_hash"`
	Sig1       string      `json:"sig1"`
	Sig2       string      `json:"sig2"`
}

func NewFuzzyHash(InputLen int) *FuzzyHash {
	if InputLen < 1 {
		return nil
	}

	fh := new(FuzzyHash)
	fh.InputLen = InputLen
	fh.Retry = true
	fh.Bs = calcInitBlockSize(uint32(InputLen))
	fh.resetFH()
	return fh
}

func (fh *FuzzyHash) Step(d byte) {
	fh.Index += 1
	if fh.Index >= fh.InputLen {
		fh.Sig1 += string(B64_CHARS[fh.Hash1.Sum32()&0x3F])
		fh.Sig2 += string(B64_CHARS[fh.Hash2.Sum32()&0x3F])

		if uint32(len(fh.Sig1)) >= SS_LEN/2 || fh.Bs == blockSizeMin {
			fh.Retry = false
			return
		}

		fh.resetFH()
		fh.Bs = fh.Bs / 2
		return
	}

	rs := fh.Rh.hash(d)
	fh.Hash1.Write([]byte{d})
	fh.Hash2.Write([]byte{d})
	fh.IsTrigger1, fh.IsTrigger2 = false, false

	if mod := rs % fh.Bs; mod == fh.Bs-1 {
		fh.Sig1 += string(B64_CHARS[fh.Hash1.Sum32()&0x3F])
		fh.IsTrigger1 = true
		fh.Hash1.Reset() // reinit the hash
	}

	if mod := rs % (2 * fh.Bs); mod == (2*fh.Bs)-1 {
		fh.Sig2 += string(B64_CHARS[fh.Hash2.Sum32()&0x3F])
		fh.IsTrigger2 = true
		fh.Hash2.Reset() // reinit the hash
	}
}

func (fh *FuzzyHash) PrintSSDeep() string {
	return fmt.Sprintf("%d:%s:%s", fh.Bs, fh.Sig1, fh.Sig2)
}

func (fh *FuzzyHash) resetFH() {
	fh.Hash1, fh.Hash2 = *NewFNV(), *NewFNV()
	fh.Index = -1
	fh.Rh = *NewRollingHash()
	fh.Sig1, fh.Sig2 = "", ""
}

type RollingHash struct {
	X      uint32   `json:"x"`
	Y      uint32   `json:"y"`
	Z      uint32   `json:"z"`
	C      uint32   `json:"c"`
	Size   uint32   `json:"size"`
	Window []uint32 `json:"window"`
}

func NewRollingHash() *RollingHash {
	return &RollingHash{
		Size:   WINDOW_SIZE,
		Window: make([]uint32, WINDOW_SIZE),
	}
}

func (rh *RollingHash) hash(d byte) uint32 {
	dint := uint32(d)
	rh.Y = rh.Y - rh.X
	rh.Y = rh.Y + rh.Size*dint
	rh.X = rh.X + dint
	rh.X = rh.X - rh.Window[rh.C%rh.Size]
	rh.Window[rh.C%rh.Size] = dint
	rh.C = rh.C + 1
	rh.Z = rh.Z << 5
	rh.Z = rh.Z ^ dint

	return (rh.X + rh.Y + rh.Z)
}
