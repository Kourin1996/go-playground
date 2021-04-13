package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
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
	tmp := make([]byte, 64)
	tmp[0] = 0x80

	// 64 bytes block
	// (original message) + (0x80) + (k bytes 0 padding) + (8bytes original length)
	// k = 56 - 1 - l
	if l%64 <= 55 {
		input = append(input, tmp[0:56-l%64]...)
	} else {
		input = append(input, tmp[0:64+56-l%64]...)
	}

	// add original message size (bits)
	binary.BigEndian.PutUint64(tmp[:], l<<3)
	return append(input, tmp[0:8]...)
}

func decomposition(input []byte) []byte {
	l := len(input)
	blockSize := l / 64
	output := make([]byte, 256*blockSize)

	for b := 0; b < blockSize; b++ {
		// original => 512 bits, 64 byte
		// w => 2048 bits, 256 byte

		// W1,W2,...,W16 = M[b]
		copy(output[b*256:b*256+64], input[b*64:(b+1)*64])
		for k := 16; k < 64; k++ {
			i := b*256 + k*4

			wim2 := binary.BigEndian.Uint32(output[i-4*2 : i-4*1])
			wim7 := binary.BigEndian.Uint32(output[i-4*7 : i-4*6])
			wim15 := binary.BigEndian.Uint32(output[i-4*15 : i-4*14])
			wim16 := binary.BigEndian.Uint32(output[i-4*16 : i-4*15])

			x1 := bits.RotateLeft32(wim2, -17) ^ bits.RotateLeft32(wim2, -19) ^ (wim2 >> 10)
			x2 := wim7
			x3 := bits.RotateLeft32(wim15, -7) ^ bits.RotateLeft32(wim15, -18) ^ (wim15 >> 3)
			x4 := wim16
			binary.BigEndian.PutUint32(output[i:i+4], x1+x2+x3+x4)
		}
	}
	return output
}

func hash(input []byte) []byte {
	l := len(input)
	blockSize := l / 256

	hash := make([]byte, 32)
	for i := 0; i < 8; i++ {
		binary.BigEndian.PutUint32(hash[i*4:(i+1)*4], InitHash[i])
	}

	for b := 0; b < blockSize; b++ {
		var a2h [8]uint32
		for i := 0; i < 8; i++ {
			a2h[i] = binary.BigEndian.Uint32(hash[i*4 : (i+1)*4])
		}

		for i := 0; i < 64; i++ {
			t11 := a2h[7]
			t12 := bits.RotateLeft32(a2h[4], -6) ^ bits.RotateLeft32(a2h[4], -11) ^ bits.RotateLeft32(a2h[4], -25)
			t13 := (a2h[4] & a2h[5]) ^ ((^a2h[4]) & a2h[6])
			t14 := KArray[i]
			t15 := binary.BigEndian.Uint32(input[b*256+i*4 : b*256+(i+1)*4])
			t21 := bits.RotateLeft32(a2h[0], -2) ^ bits.RotateLeft32(a2h[0], -13) ^ bits.RotateLeft32(a2h[0], -22)
			t22 := (a2h[0] & a2h[1]) ^ (a2h[0] & a2h[2]) ^ (a2h[1] & a2h[2])
			t1 := t11 + t12 + t13 + t14 + t15
			t2 := t21 + t22

			a2h[7] = a2h[6]      // h = g
			a2h[6] = a2h[5]      // g = f
			a2h[5] = a2h[4]      // f = e
			a2h[4] = a2h[3] + t1 // e = d+T1
			a2h[3] = a2h[2]      // d = c
			a2h[2] = a2h[1]      // c = b
			a2h[1] = a2h[0]      // b = a
			a2h[0] = t1 + t2     // a = t1 + t2
		}

		// update hash array
		for i := 0; i < 8; i++ {
			oldH := binary.BigEndian.Uint32(hash[i*4 : (i+1)*4])
			newH := oldH + a2h[i]
			binary.BigEndian.PutUint32(hash[i*4:(i+1)*4], newH)
		}
	}
	return hash
}

func sha256(input []byte) ([]byte, error) {
	result := make([]byte, len(input))
	copy(result, input[:])
	fmt.Printf("original => %s\n", getHex(result))
	result = padding(result)
	fmt.Printf("padding => %s\n", getHex(result))
	result = decomposition(result)
	fmt.Printf("decomposition => %s\n", getHex(result))
	result = hash(result)
	fmt.Printf("hash => %s\n", getHex(result))
	return result, nil
}

func getHex(bytes []byte) string {
	// hash := make([]byte, hex.EncodedLen(len(bytes)))
	// hex.Encode(hash, bytes)
	// return string(hash)
	return fmt.Sprintf("%08b", bytes)
}

func main() {
	msg := "hello world"
	hashBytes, err := sha256([]byte(msg))
	if err != nil {
		log.Fatal(err)
	}
	hash := make([]byte, hex.EncodedLen(len(hashBytes)))
	hex.Encode(hash, hashBytes)
	fmt.Printf("result: %s\n", hash)
}
