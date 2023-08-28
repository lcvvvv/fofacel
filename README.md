# fofacel
一个Fofa语法的表达式解析库，用于检测关键字是否满足满足Fofa表达式

# 示例

```go
package main

import (
	"fmt"
	"github.com/lcvvvv/fofacel"
)

func main() {
	engine := fofacel.New("body","title")
	
	
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
```

