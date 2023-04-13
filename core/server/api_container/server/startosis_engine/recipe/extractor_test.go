package recipe

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.starlark.net/starlark"
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
		result, err := Extractor(queryList[i], jsonObject)
		assert.Nil(t, err)
		assert.Equal(t, expectedList[i], result)
	}
}

func TestExtractor_FailureQueryNotFound(t *testing.T) {
	queryList := []string{".not_found", ".list.[2]", ".dict.not_found"}
	for i := range queryList {
		result, err := Extractor(queryList[i], jsonObject)
		assert.Nil(t, result)
		assert.NotNil(t, err)
	}
}

func TestExtractor_FailureQueryInvalid(t *testing.T) {
	notValidQuery := "not valid query"
	result, err := Extractor(notValidQuery, jsonObject)
	assert.Nil(t, result)
	assert.NotNil(t, err)
}

func TestExtractor_FailureInput(t *testing.T) {
	badInput := []byte(`{
"key": "value"`)
	trivialQuery := "."
	result, err := Extractor(trivialQuery, badInput)
	assert.Nil(t, result)
	assert.NotNil(t, err)
}
