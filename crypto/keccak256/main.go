package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
)

type Lane uint64

type State [5][5]Lane

var L = [7]uint8{0, 1, 2, 3, 4, 5, 6}
var N = [7]uint8{}
var W = [7]uint8{}
var B = [7]uint16{}

var RC = [24]uint64{
	0x0000000000000001, 0x0000000000008082, 0x800000000000808A, 0x8000000080008000,
	0x000000000000808B, 0x0000000080000001, 0x8000000080008081, 0x8000000000008009,
	0x000000000000008A, 0x0000000000000088, 0x0000000080008009, 0x000000008000000A,
	0x000000008000808B, 0x800000000000008B, 0x8000000000008089, 0x8000000000008003,
	0x8000000000008002, 0x8000000000000080, 0x000000000000800A, 0x800000008000000A,
	0x8000000080008081, 0x8000000000008080, 0x0000000080000001, 0x8000000080008008,
}

var R = [5][5]uint8{
	{0, 36, 3, 41, 18},
	{1, 44, 10, 45, 2},
	{62, 6, 43, 15, 61},
	{28, 55, 25, 21, 56},
	{27, 20, 39, 8, 14},
}

func init() {
	for i, l := range L {
		N[i] = 12 + 2*l
		W[i] = uint8(math.Pow(float64(2), float64(l)))
		B[i] = uint16(25) * uint16(W[i])
	}
}

func findB(b uint16) (index int) {
	index = -1
	for i, x := range B {
		if b == x {
			index = i
			break
		}
	}
	return
}

func getLaneSize(r uint16, c uint16) uint8 {
	i := findB(r + c)
	if i == -1 {
		panic("b not found")
	}
	return W[i]
}

func mod(x int, m int) int {
	if x < 0 {
		return x + m
	}
	return x % m
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func rot(x uint64, n int8, s uint8) uint64 {
	if n == 0 {
		return x
	}
	isLeft, an := n > 0, uint8(mod(abs(int(n)), int(s)))
	mask := ^(uint64(math.Pow(2, float64(64-s))) - 1)
	if isLeft {
		return (mask & ((x & mask) << an)) | (mask & ((x & mask) >> (s - an)))
	} else {
		return (mask & ((x & mask) >> an)) | (mask & ((x & mask) << (s - an)))
	}
}

func round(b uint16, A *State, rc uint64, w uint8) {
	var C, D [5]Lane
	var B [5][5]Lane

	// θ step
	for x := 0; x < 5; x++ {
		C[x] = A[x][0] ^ A[x][1] ^ A[x][2] ^ A[x][3] ^ A[x][4]
	}
	for x := 0; x < 5; x++ {
		xm1, xm2 := mod(x-1, 5), mod(x+1, 5)
		D[x] = Lane(uint64(C[xm1]) ^ rot(uint64(C[xm2]), -1, w))
	}
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			A[x][y] ^= D[x]
		}
	}

	// ρ and π steps
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			bx, by := y, mod(2*x+3*y, 5)
			B[bx][by] = Lane(rot(uint64(A[x][y]), int8(R[x][y]), w))
		}
	}

	// χ step
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			a, b, c := uint64(B[x][y]), uint64(B[mod(x+1, 5)][y]), uint64(B[mod(x+2, 5)][y])
			A[x][y] = Lane(a ^ (^b & c))
		}
	}

	// ι step
	A[0][0] = Lane(uint64(A[0][0]) ^ rc)
}

func keccakf(b uint16, A *State) {
	bi := findB(b)
	if bi == -1 {
		panic("b not found")
	}
	n, w := N[bi], W[bi]

	for i := uint8(0); i < n; i++ {
		rc := RC[i]
		round(b, A, rc, w)
	}
}

func keccak(r uint16, c uint16, input []byte, d uint8, outputLen uint64) []byte {
	if ((r + c) != 1600) || ((r % 8) != 0) {
		panic("rate and capacity are invalid")
	}
	rateInBytes := int(r / 8)
	laneSize, state := getLaneSize(r, c), new(State)

	fmt.Printf("r=%d, c=%d, d=%b, laneSize=%d\n", r, c, d, laneSize)

	blockSize, inputLen := 0, len(input)
	for inputLen > 0 {
		blockSize = inputLen
		if inputLen > rateInBytes {
			blockSize = rateInBytes
		}

		my := int(blockSize / 5)
		for y := 0; y < my; y++ {
			mx := blockSize - 5*y
			if mx < 0 {
				mx = 0
			}

			for x := 0; x < mx; x++ {
				state[mod(x, 5)][mod(y, 5)] ^= Lane(input[x+5*y])
			}
		}

		inputLen -= blockSize
		if blockSize == rateInBytes {
			keccakf(r+c, state)
			blockSize = 0
		}
	}

	state[blockSize%5][blockSize/5] ^= Lane(d)

	if (d&0x80) != 0 && blockSize == (rateInBytes-1) {
		keccakf(r+c, state)
	}

	state[(rateInBytes-1)%5][mod((rateInBytes-1)/5, 5)] ^= 0x80
	keccakf(r+c, state)

	output := make([]byte, 0, outputLen)
	for outputLen > 0 {
		blockSize = int(outputLen)
		if blockSize > rateInBytes {
			blockSize = rateInBytes
		}
		fmt.Printf("outputLen=%d, rateInBytes=%d, blockSize=%d[bytes]\n", outputLen, rateInBytes, blockSize)

		// my := int(int(r) / (int(laneSize) * 5))
		// for y := 0; y < my; y++ {
		// 	mx := int(r)/int(laneSize) - 5*y
		// 	if mx < 0 {
		// 		mx = 0
		// 	}

		// 	for x := 0; x < mx; x++ {
		// 		s := state[mod(x, 5)][mod(y, 5)]
		// 		bytes := make([]byte, 8)
		// 		binary.LittleEndian.PutUint64(bytes, uint64(s))
		// 		output = append(output, bytes...)
		// 	}
		// }

		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				s := state[mod(x, 5)][mod(y, 5)]
				bytes := make([]byte, 8)
				binary.LittleEndian.PutUint64(bytes, uint64(s))
				output = append(output, bytes...)
			}
		}

		outputLen -= uint64(blockSize)
		if outputLen > 0 {
			keccakf(r+c, state)
		}
	}

	return output
}

func SHAKE128(input []byte, outputLen uint64) []byte {
	return keccak(1344, 256, input, 0x1f, outputLen)
}

func SHAKE256(input []byte, outputLen uint64) []byte {
	return keccak(1088, 512, input, 0x1f, outputLen)
}

func SHA3_224(input []byte) []byte {
	return keccak(1152, 448, input, 0x06, 224/8)
}

func SHA3_256(input []byte) []byte {
	return keccak(1088, 512, input, 0x06, 256/8)
}

func SHA3_384(input []byte) []byte {
	return keccak(832, 768, input, 0x06, 384/8)
}

func SHA3_512(input []byte) []byte {
	return keccak(576, 1024, input, 0x06, 512/8)
}

func getHex(bytes []byte) string {
	hash := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(hash, bytes)
	return string(hash)
}

func main() {
	msg := "hello world"
	res := SHA3_224([]byte(msg))
	fmt.Println(getHex(res))
}
