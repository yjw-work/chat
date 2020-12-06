package main

import (
	"fmt"
	"testing"
)

func TestWordFilt(t *testing.T) {
	filter := NewTrie()
	filter.Insert("abc")
	fmt.Println(filter.Filt("abc"))
	fmt.Println(filter.Filt("a"))
	fmt.Println(filter.Filt("b"))
	fmt.Println(filter.Filt("c"))
	fmt.Println(filter.Filt("ab"))
	fmt.Println(filter.Filt("bc"))
	fmt.Println(filter.Filt("abcd"))
}
