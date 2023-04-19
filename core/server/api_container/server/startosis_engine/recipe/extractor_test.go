package recipe

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
	"testing"
)

var jsonObject = []byte(`
{
	"integer": 1,
	"str": "a",
	"list": [0.2, "b"],
	"dict": {
		"key": "value",
		"another_key": -1
	},
	"bool": false
}
`)

func TestExtractor_Success(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	expectedDict := starlark.NewDict(2)
	_ = expectedDict.SetKey(starlark.String("key"), starlark.String("value"))
	_ = expectedDict.SetKey(starlark.String("another_key"), starlark.MakeInt(-1))
	queryList := []string{".integer", ".str", ".list", ".dict.key", ".bool", ".dict", ".integer, .str"}
	expectedList := []starlark.Comparable{
		starlark.MakeInt(1),
		starlark.String("a"),
		starlark.NewList([]starlark.Value{starlark.Float(0.2), starlark.String("b")}),
		starlark.String("value"),
		starlark.Bool(false),
		expectedDict,
		starlark.NewList([]starlark.Value{starlark.MakeInt(1), starlark.String("a")}),
	}
	for i := range queryList {
		result, extractErr := extract(jsonObject, queryList[i])
		assert.Nil(t, extractErr)
		equal, cmpErr := expectedList[i].CompareSameType(syntax.EQL, result, 2)
		assert.Nil(t, cmpErr)
		assert.True(t, equal)
	}
}

func TestExtractor_FailureQueryNotFound(t *testing.T) {
	queryList := []string{".not_found", ".list.[2]", ".dict.not_found"}
	for i := range queryList {
		result, err := extract(jsonObject, queryList[i])
		assert.Nil(t, result)
		assert.NotNil(t, err)
	}
}

func TestExtractor_FailureQueryInvalid(t *testing.T) {
	notValidQuery := "not valid query"
	result, err := extract(jsonObject, notValidQuery)
	assert.Nil(t, result)
	assert.NotNil(t, err)
}

func TestExtractor_FailureInput(t *testing.T) {
	testInputs := map[string]string{
		"1":       ".",
		"\"hi\"":  "length",
		"\"bye\"": ".",
	}
	testOutputs := map[string]starlark.Comparable{
		"1":       starlark.MakeInt(1),
		"\"hi\"":  starlark.MakeInt(2),
		"\"bye\"": starlark.String("bye"),
	}
	index := 0
	for input, query := range testInputs {
		result, err := extract([]byte(input), query)
		assert.Nil(t, err)
		assert.Equal(t, testOutputs[input], result)
		index += 1
	}
}

func TestExtractor_SimpleValues(t *testing.T) {
	extractors := map[string]string{
		"id":  ".integer",
		"id2": ".str",
	}
	result, err := runExtractors(jsonObject, extractors)
	assert.Nil(t, err)
	assert.Equal(t, map[string]starlark.Comparable{
		"extract.id":  starlark.MakeInt(1),
		"extract.id2": starlark.String("a"),
	}, result)
}

func TestRunExtractors_OutputKeys(t *testing.T) {
	extractors := map[string]string{
		"id":  ".integer",
		"id2": ".str",
	}
	result, err := runExtractors(jsonObject, extractors)
	assert.Nil(t, err)
	assert.Equal(t, map[string]starlark.Comparable{
		"extract.id":  starlark.MakeInt(1),
		"extract.id2": starlark.String("a"),
	}, result)
}

func TestRunExtractors_EmptyOutput(t *testing.T) {
	extractors := map[string]string{}
	result, err := runExtractors(jsonObject, extractors)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestRunExtractors_FailureQueryInvalid(t *testing.T) {
	extractors := map[string]string{
		"id": ".does_not_exist",
	}
	result, err := runExtractors(jsonObject, extractors)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}
