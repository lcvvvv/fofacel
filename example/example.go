package main

import (
	"fmt"
	"github.com/lcvvvv/fofacel"
)

func main() {
	rule, err := fofacel.New(`body="aaaaa" && title="aaaaaa"`)
	if err != nil {
		panic(err)
	}

	fmt.Println(rule.Match(fofacel.NewKeywords(map[string]string{
		"body":  "aaaaaaaaaaaaaaaa",
		"title": "aaaaaaaaaaaaaa",
	})))
	//true

	fmt.Println(rule.Match(fofacel.NewKeywords(map[string]string{
		"body":  "bbbbbbbbbbbbb",
		"title": "aaaaaaaaaaaaaa",
	})))
	//false
}
