package prefab

import (
	"fmt"
	"math"

	"github.com/spaolacci/murmur3"
)

type Hashing struct{}

// HashZeroToOne
func (Hashing) HashZeroToOne(value string) (hashedValue float64, ok bool) {
	h32 := murmur3.New32()
	_, err := h32.Write([]byte(value))
	if err != nil {
		fmt.Println("Error hashing data:", err)
		return 0, false
	}
	return float64(h32.Sum32()) / float64(math.MaxUint32), true
}
