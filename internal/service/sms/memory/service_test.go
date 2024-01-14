package memory

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestService_Send(t *testing.T) {
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	fmt.Println(code)
}
