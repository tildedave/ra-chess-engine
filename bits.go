package main

/*
typedef unsigned long long uint64_t;

int _ctzll(uint64_t x)
{
    return __builtin_ctzll(x);
}

int _popcount(uint64_t x)
{
	return __builtin_popcountll(x);
}

*/
import "C"

func TrailingZeros64(x uint64) int {
	return int(C._ctzll(C.ulonglong(x)))
}

func OnesCount64(x uint64) int {
	return int(C._popcount(C.ulonglong(x)))
}
