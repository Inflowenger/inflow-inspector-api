package main

import (
	"fmt"
	"testing"

	"github.com/Inflowenger/inflow-inspector-api/repository"
)

func TestXxx(t *testing.T) {
	lk, _ := repository.GetAllKeys()
	for _, el := range lk {
		fmt.Println(el)
	}
}
