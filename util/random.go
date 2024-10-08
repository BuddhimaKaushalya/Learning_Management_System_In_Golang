package util

import (
	"math/rand"
	"strings"
	"time"
)

const digits = "1234567890"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomInt generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(digits)

	for i := 0; i < n; i++ {
		c := digits[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

//Random OTP with 6 digits

// RandomOwner generates a random owner name
func RandomOwner() string {
	return RandomString(6)
}
