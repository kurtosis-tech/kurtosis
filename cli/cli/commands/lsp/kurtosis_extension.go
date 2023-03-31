package lsp

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/vscode-kurtosis/starlark-lsp/pkg/analysis"
	"github.com/kurtosis-tech/vscode-kurtosis/starlark-lsp/pkg/docstring"
	"github.com/kurtosis-tech/vscode-kurtosis/starlark-lsp/pkg/query"
	"github.com/sirupsen/logrus"
	"sync"
)

//go:embed resource/kurtosis_starlark.json
var kurtosisStarlarkJson []byte
var once sync.Once

type KurtosisExtensionWrapper struct {
	symbols    []query.Symbol
	signatures map[string]query.Signature
	members    []query.Symbol
	methods    map[string]query.Signature
}

type KurtosisExtensionContract struct {
	Name          string `json:"name"`
	Detail        string `json:"detail"`
	Documentation string `json:"documentation"`
	ReturnType    string `json:"returnType"`
	Params        []struct {
		Name         string `json:"name"`
		Type         string `json:"type"`
		Content      string `json:"content"`
		DefaultValue string `json:"defaultValue"`
		Detail       string `json:"detail"`
	} `json:"params"`
}

type KurtosisBuiltins struct {
	TypeBuiltIns   []KurtosisExtensionContract `json:"type_builtins"`
	MethodBuiltIns []KurtosisExtensionContract `json:"method_builtins"`
}

type KurtosisBuiltinProvider struct {
	kurtosisBuiltins KurtosisBuiltins
}

// this method is the main method which converts our definition into types recognized by starlark lsp
// TODO: will make this more readable in future
func (definition *KurtosisBuiltinProvider) convertJsonToKurotisWrapper() KurtosisExtensionWrapper {
	var symbols []query.Symbol
	signatures := make(map[string]query.Signature)

	for _, plugin := range definition.kurtosisBuiltins.TypeBuiltIns {
		// query symbol
		// nolint:exhaustruct
		symbol := query.Symbol{
			Name:          plugin.Name,
			Detail:        plugin.Detail,
			Documentation: plugin.Documentation,
			KType:         true,
		}

		var signatureParams []query.Parameter
		var docStringFields []docstring.Field

		for _, param := range plugin.Params {
			// nolint:exhaustruct
			signatureParam := query.Parameter{
				Name:         param.Name,
				TypeHint:     param.Type,
				DefaultValue: param.DefaultValue,
				Content:      fmt.Sprintf("%v:%v", param.Content, param.Type),
			}

			signatureParams = append(signatureParams, signatureParam)
			docstringField := docstring.Field{Name: param.Name, Desc: param.Detail}
			docStringFields = append(docStringFields, docstringField)
		}

		// nolint:exhaustruct
		parsed := docstring.Parsed{
			Description: "",
			Fields: []docstring.FieldsBlock{
				{
					Title:  "Args",
					Fields: docStringFields,
				},
			},
		}

		// nolint:exhaustruct
		signature := query.Signature{
			Name:       plugin.Name,
			Params:     signatureParams,
			ReturnType: plugin.ReturnType,
			Docs:       parsed,
		}
		symbols = append(symbols, symbol)
		signatures[plugin.Name] = signature
	}

	var members []query.Symbol
	methods := make(map[string]query.Signature)
	for _, plugin := range definition.kurtosisBuiltins.MethodBuiltIns {
		// query symbol
		// nolint:exhaustruct
		member := query.Symbol{
			Name:          plugin.Name,
			Detail:        plugin.Detail,
			Documentation: plugin.Documentation,
			KType:         true,
		}

		var signatureParams []query.Parameter
		var docStringFields []docstring.Field

		for _, param := range plugin.Params {
			// nolint:exhaustruct
			signatureParam := query.Parameter{
				Name:         param.Name,
				TypeHint:     param.Type,
				DefaultValue: param.DefaultValue,
				Content:      fmt.Sprintf("%v:%v", param.Content, param.Type),
			}

			signatureParams = append(signatureParams, signatureParam)
			docstringField := docstring.Field{Name: param.Name, Desc: param.Detail}
			docStringFields = append(docStringFields, docstringField)
		}

		// nolint:exhaustruct
		parsed := docstring.Parsed{
			Description: "",
			Fields: []docstring.FieldsBlock{
				{
					Title:  "Args",
					Fields: docStringFields,
				},
			},
		}

		// nolint:exhaustruct
		method := query.Signature{
			Name:       plugin.Name,
			Params:     signatureParams,
			ReturnType: plugin.ReturnType,
			Docs:       parsed,
		}
		members = append(members, member)
		methods[plugin.Name] = method
	}

	return KurtosisExtensionWrapper{
		symbols:    symbols,
		signatures: signatures,
		methods:    methods,
		members:    members,
	}
}

func (definition *KurtosisBuiltinProvider) ReadJsonFile() error {
	var kurtosisBuiltIns KurtosisBuiltins
	var err error

	once.Do(func() {
		err = json.Unmarshal(kurtosisStarlarkJson, &kurtosisBuiltIns)
		if err == nil {
			definition.kurtosisBuiltins = kurtosisBuiltIns
		}
	})
	return err
}

func GetKurtosisBuiltIn() *analysis.Builtins {
	// nolint:exhaustruct
	kurtosisProvider := KurtosisBuiltinProvider{}
	builtIn := analysis.NewBuiltins()

	err := kurtosisProvider.ReadJsonFile()

	// silently logging the failure so that lsp server still works even without kurtosis builtins
	if err != nil {
		logrus.Debugf("Error occurred while getting kurtosis builtins - %+v", err)
		return builtIn
	}

	converted := kurtosisProvider.convertJsonToKurotisWrapper()
	builtIn.Symbols = converted.symbols
	builtIn.Functions = converted.signatures
	builtIn.Methods = converted.methods
	builtIn.Members = converted.members
	return builtIn
}
