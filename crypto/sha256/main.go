package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/bits"
)

var InitHash [8]uint32 = [8]uint32{
	0x6A09E667,
	0xBB67AE85,
	0x3C6EF372,
	0xA54FF53A,
	0x510E527F,
	0x9B05688C,
	0x1F83D9AB,
	0x5BE0CD19,
}

var KArray [64]uint32 = [64]uint32{
	0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
	0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
	0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
	0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
	0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
	0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
	0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
	0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
}

func padding(input []byte) []byte {
	l := uint64(len(input))
	// 64 bytes block
	// (original message) + (0x80) + (k bytes 0 padding) + (8bytes original length)
	// k = 56 - 1 - l
	newSize := l + 1 + 8 + (55 - l%64)
	if l%64 >= 55 {
		newSize += 64
	}

	output := make([]byte, newSize)
	copy(output[:l], input)
	output[l] = 0x80
	// add original message size (bits)
	binary.BigEndian.PutUint64(output[newSize-8:newSize], l<<3)

	return output
}

func bytes2Uint32Array(input []byte) []uint32 {
	size := len(input) / 4
	output := make([]uint32, size)
	for i := 0; i < size; i++ {
		for k := 0; k < 4; k++ {
			output[i] |= uint32(input[i*4+k]) << ((3 - k) * 8)
		}
	}
	return output
}

func uint32Array2Bytes(input []uint32) []byte {
	size := len(input) * 4
	output := make([]byte, size)
	for i := 0; i < size; i++ {
		output[i] = byte(input[i/4] >> (8 * (3 - i%4)))
	}
	return output
}

func hash(input []uint32) [8]uint32 {
	l := len(input)
	blockSize := l / 16

	var hash [8]uint32
	var a2h [8]uint32
	var w [64]uint32

	copy(hash[:], InitHash[:])
	copy(a2h[:], hash[:])

	for b := 0; b < blockSize; b++ {
		// decomposition
		// original block => 512 bits, 64 byte
		// w => 2048 bits, 256 byte
		// W1,W2,...,W16 = M[b]
		for i := 0; i < 64; i++ {
			if i < 16 {
				w[i] = input[b*16+i]
			} else {
				t1 := bits.RotateLeft32(w[i-2], -17) ^ bits.RotateLeft32(w[i-2], -19) ^ (w[i-2] >> 10)
				t2 := w[i-7]
				t3 := bits.RotateLeft32(w[i-15], -7) ^ bits.RotateLeft32(w[i-15], -18) ^ (w[i-15] >> 3)
				t4 := w[i-16]
				w[i] = t1 + t2 + t3 + t4
			}
		}

		// hash
		for i := 0; i < 64; i++ {
			t11 := a2h[7]
			t12 := bits.RotateLeft32(a2h[4], -6) ^ bits.RotateLeft32(a2h[4], -11) ^ bits.RotateLeft32(a2h[4], -25)
			t13 := (a2h[4] & a2h[5]) ^ ((^a2h[4]) & a2h[6])
			t14 := KArray[i]
			t15 := w[i]
			t21 := bits.RotateLeft32(a2h[0], -2) ^ bits.RotateLeft32(a2h[0], -13) ^ bits.RotateLeft32(a2h[0], -22)
			t22 := (a2h[0] & a2h[1]) ^ (a2h[0] & a2h[2]) ^ (a2h[1] & a2h[2])
			t1 := t11 + t12 + t13 + t14 + t15
			t2 := t21 + t22
			a2h[7], a2h[6], a2h[5], a2h[4], a2h[3], a2h[2], a2h[1], a2h[0] =
				a2h[6], a2h[5], a2h[4], a2h[3]+t1, a2h[2], a2h[1], a2h[0], t1+t2
		}
		for i := 0; i < 8; i++ {
			hash[i] += a2h[i]
			a2h[i] = hash[i]
		}
	}
	return hash
}

func sha256(input []byte) []byte {
	r1 := padding(input)
	r2 := bytes2Uint32Array(r1)
	r3 := hash(r2)
	return uint32Array2Bytes(r3[:])
}

func getHex(bytes []byte) string {
	hash := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(hash, bytes)
	return string(hash)
}

func main() {
	msg := "hello world"
	res := sha256([]byte(msg))
	fmt.Println(getHex(res)) // => b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
}
