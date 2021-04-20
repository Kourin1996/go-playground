package main

// WIP: This is not finished to implement

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
)

type State []byte

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	isRight, an := n > 0, uint8(mod(abs(int(n)), int(s)))
	mask := ^(uint64(math.Pow(2, float64(64-s))) - 1)
	if isRight {
		return (mask & ((x & mask) >> an)) | (mask & ((x & mask) << (s - an)))
	} else {
		return (mask & ((x & mask) << an)) | (mask & ((x & mask) >> (s - an)))
	}
}

func getBeginLaneIndex(l uint8, x, y int) int {
	return (x + 5*y) * int(l) / 8
}

func readLane(A []byte, l uint8, x, y int) uint64 {
	begin := getBeginLaneIndex(l, x, y)
	end := begin + int(l)/8
	return binary.BigEndian.Uint64(A[begin:end])
}

func writeLane(A []byte, l uint8, x, y int, v uint64) {
	offset := getBeginLaneIndex(l, x, y)
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, v)
	copy(A[offset:offset+int(l)/8], bytes[0:int(l)/8])
}

func xorLane(A []byte, l uint8, x, y int, v uint64) {
	offset := getBeginLaneIndex(l, x, y)
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, v)
	for i := uint8(0); i < l/8; i++ {
		A[offset+int(i)] ^= bytes[i]
	}
}

func round(b uint16, A []byte, rc uint64, w uint8) {
	var C [5]uint64
	var D uint64

	// θ step
	for x := 0; x < 5; x++ {
		C[x] = readLane(A, w, x, 0) ^ readLane(A, w, x, 1) ^ readLane(A, w, x, 2) ^ readLane(A, w, x, 3) ^ readLane(A, w, x, 4)
	}
	for x := 0; x < 5; x++ {
		xm1, xm2 := mod(x-1, 5), mod(x+1, 5)
		D = C[xm1] ^ rot(C[xm2], 1, w)
		for y := 0; y < 5; y++ {
			xorLane(A, w, x, y, D)
		}
	}
	// ρ and π steps
	{
		x, y := 1, 0
		current, temp := readLane(A, w, x, y), uint64(0)
		for t := 0; t < 24; t++ {
			r, Y := int8(((t+1)*(t+2)/2)%64), (2*x+3*y)%5
			x, y = y, Y
			temp = readLane(A, w, x, y)
			writeLane(A, w, x, y, rot(current, r, w))
			current = temp
		}
	}

	// χ step
	{
		var temp [5]uint64
		for y := 0; y < 5; y++ {
			for x := 0; x < 5; x++ {
				temp[x] = readLane(A, w, x, y)
			}
			for x := 0; x < 5; x++ {
				writeLane(A, w, x, y, temp[x]^((^temp[mod(x+1, 5)])&temp[mod(x+2, 5)]))
			}
		}
	}

	xorLane(A, w, 0, 0, rc)
}

func keccakf(b uint16, A []byte) {
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

func keccak(r uint16, c uint16, input []byte, d byte, outputLen uint64) []byte {
	if ((r + c) != 1600) || ((r % 8) != 0) {
		panic("rate and capacity are invalid")
	}
	rateInBytes, laneSize, b := int(r/8), getLaneSize(r, c), r+c
	state := make([]byte, 5*5*int(laneSize)/8)

	fmt.Printf("r=%d, c=%d, d=%b, rateInBytes=%d, laneSize=%d, outputLen=%d, stateSize=%d\n", r, c, d, rateInBytes, laneSize, outputLen, len(state))

	inputLen := len(input)
	blockSize, offset := 0, 0
	for inputLen > 0 {
		blockSize = min(inputLen, rateInBytes)
		copy(state[0:blockSize], input[offset:offset+blockSize])

		inputLen -= blockSize
		offset += blockSize
		if blockSize == rateInBytes {
			keccakf(b, state)
			blockSize = 0
		}
	}

	state[blockSize] ^= d
	if ((d & 0x80) != 0) && (blockSize == (rateInBytes - 1)) {
		keccakf(b, state)
	}
	state[rateInBytes-1] ^= 0x80
	keccakf(r+c, state)

	output, offset := make([]byte, outputLen), 0
	for outputLen > 0 {
		blockSize := min(int(outputLen), rateInBytes)
		copy(output[offset:blockSize], state[0:blockSize])
		outputLen -= uint64(blockSize)
		offset += blockSize

		if outputLen > 0 {
			keccakf(b, state)
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
	res := SHA3_256([]byte(msg))
	fmt.Println(getHex(res))
}
