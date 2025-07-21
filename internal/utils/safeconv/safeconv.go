package safeconv

import (
	"fmt"
	"math"
)

// UintToUint32 safely converts uint to uint32
func UintToUint32(v uint) (uint32, error) {
	if v > math.MaxUint32 {
		return 0, fmt.Errorf("uint overflow: %d > MaxUint32", v)
	}
	return uint32(v), nil
}

// IntToInt32 safely converts int to int32
func IntToInt32(v int) (int32, error) {
	if v > math.MaxInt32 || v < math.MinInt32 {
		return 0, fmt.Errorf("int overflow: %d not in int32 range", v)
	}
	return int32(v), nil
}

// IntToUint64 safely converts int to uint64
func IntToUint64(v int) (uint64, error) {
	if v < 0 {
		return 0, fmt.Errorf("negative int cannot be converted to uint64")
	}
	return uint64(v), nil
}
