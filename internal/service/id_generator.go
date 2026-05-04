package service

import (
	"crypto/rand"
	"encoding/binary"
)

func randomID(start int64) (int64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}

	v := int64(binary.BigEndian.Uint64(b[:]) & 0x7fffffffffffffff)
	if v < start {
		v += start
	}

	return v, nil
}
