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

func init() {
	reloadKeywordSymbols()

	//初始化comparator正则
	var comparators []string
	for comparator := range comparatorSymbols {
		comparators = append(comparators, comparator)
	}

	comparatorRegxString = strings.Join(comparators, "|")
}

func reloadKeywordSymbols() {
	//初始化declarations
	for _, keyword := range keywordSymbols {
		declarations = append(declarations, decls.NewVar(keyword, decls.String))
	}

	//初始化keyword正则
	keywordRegxString = strings.Join(keywordSymbols, "|")

	//序列化正则表达式
	regxFofaRule = regexp.MustCompile(fmt.Sprintf(`(%s)[ \t]*(%s)[ \t]*("[^"]+")`, keywordRegxString, comparatorRegxString))
}

func SetKeyword(keywords ...string) {
	keywordSymbols = keywords
	reloadKeywordSymbols()
}

func AddKeyword(keywords ...string) {
	keywordSymbols = append(keywordSymbols, keywords...)
	reloadKeywordSymbols()
}

var (
	keywordSymbols    = []string{"header", "body", "title", "icon"}
	comparatorSymbols = map[string]string{
		"==": "Equal",
		"~=": "RegexpMatch",
		"=":  "Contains",
		"!=": "NotContains",
	}
	keywordRegxString    string
	comparatorRegxString string
	declarations         []*exprpb.Decl
	regxFofaRule         *regexp.Regexp
)

type RuleChecker struct {
	cel.Program
}

// Match 只接受小写关键字，关键字清单见：keywordSymbols
func (r *RuleChecker) Match(stringMap map[string]string) bool {
	var inputs = make(map[string]interface{})
	for _, keyword := range keywordSymbols {
		inputs[keyword] = stringMap[keyword]
	}

	out, _, err := r.Program.Eval(inputs)
	if err != nil {
		panic(err)
	}
	return out.Value().(bool)
}

func New(fofaRule string) (*RuleChecker, error) {
	// 创建 CEL 环境
	env, _ := cel.NewEnv(
		//加载关键字
		cel.Declarations(declarations...),
		//加载比对函数
		containsCelFunc,
		equalCelFunc,
		notContainsCelFunc,
		regexpMatchCelFunc,
	)

	rule := ruleConvert(fofaRule)

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
	return &RuleChecker{prg}, nil
}

func ruleConvert(fofaRule string) string {
	fofaRule = strings.ReplaceAll(fofaRule, `\"`, `[quote]`)
	fofaRule = regxFofaRule.ReplaceAllStringFunc(fofaRule, func(s string) string {
		v := regxFofaRule.FindAllStringSubmatch(s, -1)[0]
		keyword := v[1]
		comparator := v[2]
		value := v[3]
		return fmt.Sprintf("%s(%s,%s)", comparatorSymbols[comparator], keyword, value)
	})
	return strings.ReplaceAll(fofaRule, `[quote]`, `\"`)
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
			return types.Bool(strings.Contains(strings.ToLower(string(v1)), strings.ToLower(string(v2))))
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
			return types.Bool(!strings.Contains(strings.ToLower(string(v1)), strings.ToLower(string(v2))))
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
