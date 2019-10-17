package main

import (
	"fmt"
	"testing"
)

func TestMakeQuery(t *testing.T) {
	result := make_query("DEVGEN2")

	queryStr := string(result.Bytes())

	fmt.Printf(queryStr)
}
