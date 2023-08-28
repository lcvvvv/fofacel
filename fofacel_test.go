package fofacel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var engine = New("body", "header", "title")

func TestNew_Contains(t *testing.T) {
	checker, err := engine.NewRule(`body="aaaaa"`)

	assert.Nil(t, err)

	assert.True(t, checker.Match(engine.NewKeywords(map[string]string{
		"body": "aaaaaaaaa",
	})))

	assert.False(t, checker.Match(engine.NewKeywords(map[string]string{
		"body": "bbbbbb",
	})))
}

func TestNew_Equal(t *testing.T) {
	checker, err := engine.NewRule(`body=="aaaa"`)
	assert.Nil(t, err)

	assert.True(t, checker.Match(engine.NewKeywords(map[string]string{
		"body": "aaaa",
	})))

	assert.False(t, checker.Match(engine.NewKeywords(map[string]string{
		"body": "bbbbbb",
	})))
}

func TestNew_NotContains(t *testing.T) {
	checker, err := engine.NewRule(`body!="aaaa"`)
	assert.Nil(t, err)

	assert.True(t, checker.Match(engine.NewKeywords(map[string]string{
		"body":    "bbbbbbbbbb",
		"aaaaaaa": "bbbbbbbbbb",
	})))

	assert.False(t, checker.Match(engine.NewKeywords(map[string]string{
		"body": "aaaaaaaaaaa",
	})))
}

func TestNew_RegexpMatch(t *testing.T) {
	checker, err := engine.NewRule(`body~="aaa.*bbbb"`)
	assert.Nil(t, err)

	assert.True(t, checker.Match(engine.NewKeywords(map[string]string{
		"body": "aaaasdfasdfu08980asudefkjnasdbbbb",
	})))

	assert.False(t, checker.Match(engine.NewKeywords(map[string]string{
		"body": "ccccccccccc",
	})))
}

func TestNew(t *testing.T) {
	checker, err := engine.NewRule(`(body="111"||header="222") && title="333" && body="4444" || title="555555"`)
	assert.Nil(t, err)

	assert.True(t, checker.Match(engine.NewKeywords(map[string]string{
		"title": "555555",
	})))

	assert.False(t, checker.Match(engine.NewKeywords(map[string]string{
		"body":   "1114444",
		"title":  "3323",
		"header": "111",
	})))

}
