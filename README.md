# fofacel
一个Fofa语法的表达式解析库，用于检测关键字是否满足满足Fofa表达式

# 示例

```
checker, err := New(`(body="111"||header="222") && title="333" && body="4444" || title="555555"`)
checker.Match(map[string]string{
"title": "555555",
})
```

