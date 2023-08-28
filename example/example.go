package main

import (
	"fmt"
	"github.com/lcvvvv/fofacel"
)

func main() {
	engine := fofacel.New("body", "title")

	rule, err := engine.NewRule(`body="aaaaa" && title="aaaaaa"`)
	if err != nil {
		panic(err)
	}

	fmt.Println(rule.Match(engine.NewKeywords(map[string]string{
		"body":  "aaaaaaaaaaaaaaaa",
		"title": "aaaaaaaaaaaaaa",
	})))
	//true

	fmt.Println(rule.Match(engine.NewKeywords(map[string]string{
		"body":  "bbbbbbbbbbbbb",
		"title": "aaaaaaaaaaaaaa",
	})))
	//false
}
