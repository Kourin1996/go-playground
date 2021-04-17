package main

import (
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

func round(b uint16, A *State, rc uint64) {
	var C, D [5]Lane
	var B [5][5]Lane

	// θ step
	for x := 0; x < 5; x++ {
		C[x] = A[x][0] ^ A[x][1] ^ A[x][2] ^ A[x][3] ^ A[x][4]
	}
	for x := 0; x < 5; x++ {
		xm1, xm2 := mod(x-1, 5), mod(x+1, 5)
		D[x] = Lane(uint64(C[xm1]) ^ rot(uint64(C[xm2]), -1, 5))
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
			B[bx][by] = Lane(rot(uint64(A[x][y]), int8(R[x][y]), 5))
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
	n := N[bi]

	for i := uint8(0); i < n; i++ {
		rc := RC[i]
		round(b, A, rc)
	}
}

func keccak() {

}

func main() {
}
