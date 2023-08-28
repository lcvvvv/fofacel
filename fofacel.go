package fofacel

import (
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"regexp"
	"strings"
)

var (
	comparatorSymbols = map[string]string{
		"==": "Equal",
		"~=": "RegexpMatch",
		"=":  "Contains",
		"!=": "NotContains",
	}
	comparatorRegxString = `==|~=|=|!=`
)

type Keywords map[string]any

func (k Keywords) Map() map[string]any {
	return k
}

type Engine struct {
	//用于存储引擎所支持的关键字
	keywordSymbols []string

	declarations []*exprpb.Decl

	keywordRegxString string

	regxFofaRule *regexp.Regexp
}

func New(keywordSymbols ...string) *Engine {
	var e = &Engine{}
	e.keywordSymbols = keywordSymbols

	//初始化declarations
	for _, keyword := range keywordSymbols {
		e.declarations = append(e.declarations, decls.NewVar(keyword, decls.String))
		e.declarations = append(e.declarations, decls.NewVar(addLowerFlag(keyword), decls.String))
	}

	//初始化keyword正则
	e.keywordRegxString = strings.Join(keywordSymbols, "|")

	//序列化正则表达式
	e.regxFofaRule = regexp.MustCompile(fmt.Sprintf(`((?i)%s)[ \t]*(%s)[ \t]*("[^"]+")`, e.keywordRegxString, comparatorRegxString))
	return e
}

func (e *Engine) NewKeywords(stringMap map[string]string) Keywords {
	var keywordMap = make(Keywords)
	for _, keyword := range e.keywordSymbols {
		keywordMap[keyword] = stringMap[keyword]
		keywordMap[addLowerFlag(keyword)] = strings.ToLower(stringMap[keyword])
	}
	return keywordMap
}

func (e *Engine) NewRule(fofaRule string) (*RuleChecker, error) {
	// 创建 CEL 环境
	env, _ := cel.NewEnv(
		//加载关键字
		cel.Declarations(e.declarations...),
		//加载比对函数
		containsCelFunc,
		equalCelFunc,
		notContainsCelFunc,
		regexpMatchCelFunc,
	)

	rule := e.ruleConvert(fofaRule)

	// 创建 CEL 表达式
	ast, issues := env.Compile(rule)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	// 创建 CEL 评估程序
	prg, err := env.Program(ast)
	if err != nil {
		return nil, err
	}
	return &RuleChecker{prg, fofaRule}, nil
}

func (e *Engine) ruleConvert(fofaRule string) string {
	fofaRule = strings.ReplaceAll(fofaRule, `\"`, `[quote]`)

	fofaRule = e.regxFofaRule.ReplaceAllStringFunc(fofaRule, func(s string) string {
		v := e.regxFofaRule.FindAllStringSubmatch(s, -1)[0]
		keyword := v[1]
		comparator := v[2]
		value := v[3]

		switch comparator {
		case "=", "!=":
			return fmt.Sprintf("%s(%s,%s)", comparatorSymbols[comparator], addLowerFlag(strings.ToLower(keyword)), value)
		default:
			return fmt.Sprintf("%s(%s,%s)", comparatorSymbols[comparator], strings.ToLower(keyword), value)
		}
	})
	return strings.ReplaceAll(fofaRule, `[quote]`, `\"`)
}

type RuleChecker struct {
	program    cel.Program
	expression string
}

// Match 只接受小写关键字，关键字清单见：keywordSymbols
func (r *RuleChecker) Match(keywords Keywords) bool {
	out, _, err := r.program.Eval(keywords.Map())
	if err != nil {
		panic(err)
	}

	return out.Value().(bool)
}

func (r *RuleChecker) String() string {
	return r.expression
}

// containsCelFunc 忽略大小写判断是否s1包含s2 comparatorSymbols : =
var containsCelFunc = cel.Function("Contains",
	cel.Overload("Contains_string_string",
		[]*cel.Type{cel.StringType, cel.StringType},
		cel.BoolType,
		cel.BinaryBinding(func(lhs ref.Val, rhs ref.Val) ref.Val {
			v1, ok := lhs.(types.String)
			if !ok {
				return types.ValOrErr(lhs, "unexpected type '%v' passed to contains", lhs.Type())
			}
			v2, ok := rhs.(types.String)
			if !ok {
				return types.ValOrErr(rhs, "unexpected type '%v' passed to contains", rhs.Type())
			}
			return types.Bool(strings.Contains(string(v1), strings.ToLower(string(v2))))
		}),
	),
)

// equalCelFunc 判断s1与s2是否相等 comparatorSymbols : ==
var equalCelFunc = cel.Function("Equal",
	cel.Overload("Equal_string_string",
		[]*cel.Type{cel.StringType, cel.StringType},
		cel.BoolType,
		cel.BinaryBinding(func(lhs ref.Val, rhs ref.Val) ref.Val {
			v1, ok := lhs.(types.String)
			if !ok {
				return types.ValOrErr(lhs, "unexpected type '%v' passed to equal", lhs.Type())
			}
			v2, ok := rhs.(types.String)
			if !ok {
				return types.ValOrErr(rhs, "unexpected type '%v' passed to equal", rhs.Type())
			}
			return types.Bool(string(v1) == string(v2))
		}),
	),
)

// notContainsCelFunc 忽略大小写判断是否s1不包含s2 comparatorSymbols : !=
var notContainsCelFunc = cel.Function("NotContains",
	cel.Overload("NotContains_string_string",
		[]*cel.Type{cel.StringType, cel.StringType},
		cel.BoolType,
		cel.BinaryBinding(func(lhs ref.Val, rhs ref.Val) ref.Val {
			v1, ok := lhs.(types.String)
			if !ok {
				return types.ValOrErr(lhs, "unexpected type '%v' passed to notContains", lhs.Type())
			}
			v2, ok := rhs.(types.String)
			if !ok {
				return types.ValOrErr(rhs, "unexpected type '%v' passed to notContains", rhs.Type())
			}
			return types.Bool(!strings.Contains(string(v1), strings.ToLower(string(v2))))
		}),
	),
)

// regexpMatchCelFunc 判断s1是否满足正则表达式s2 comparatorSymbols : ~=
var regexpMatchCelFunc = cel.Function("RegexpMatch",
	cel.Overload("RegexpMatch_string_string",
		[]*cel.Type{cel.StringType, cel.StringType},
		cel.BoolType,
		cel.BinaryBinding(func(lhs ref.Val, rhs ref.Val) ref.Val {
			v1, ok := lhs.(types.String)
			if !ok {
				return types.ValOrErr(lhs, "unexpected type '%v' passed to regexpMatch", lhs.Type())
			}

			v2, ok := rhs.(types.String)
			if !ok {
				return types.ValOrErr(rhs, "unexpected type '%v' passed to regexpMatch", rhs.Type())
			}

			re, err := regexp.Compile(string(v2))
			if !ok {
				return types.NewErr("uncompleted value '%v", err)
			}
			return types.Bool(re.MatchString(string(v1)))
		}),
	),
)

func addLowerFlag(s string) string {
	return "ToLower" + s
}
