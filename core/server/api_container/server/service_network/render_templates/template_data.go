package render_templates

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	templateNamePrefix                   = "kurtosis-template-"
	folderPermissionForRenderedTemplates = 0755
)

type TemplateData struct {
	template *template.Template

	dataAsSerializedJson string
}

func CreateTemplateData(templateString string, dataAsSerializedJson string) (*TemplateData, error) {
	templateName, err := generateUniqueTemplateName(templateString)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error generating unique template name")
	}
	parsedTemplate, err := template.New(templateName).Parse(templateString)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing the template string '%s'", templateString)
	}

	return &TemplateData{
		template:             parsedTemplate,
		dataAsSerializedJson: dataAsSerializedJson,
	}, nil
}

func (templateData *TemplateData) GetDataAsSerializedJson() string {
	return templateData.dataAsSerializedJson
}

func (templateData *TemplateData) ReplaceRuntimeValues(runtimeValueStore *runtime_value_store.RuntimeValueStore) error {
	dataAsSerializedJsonWithRuntimeValues, err := magic_string_helper.ReplaceRuntimeValueInString(templateData.dataAsSerializedJson, runtimeValueStore)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred while replacing runtime values in JSON data '%s'",
			templateData.dataAsSerializedJson,
		)
	}
	templateData.dataAsSerializedJson = dataAsSerializedJsonWithRuntimeValues
	return nil
}

func (templateData *TemplateData) RenderToFile(destinationAbsoluteFilePath string) error {
	decodedData, err := decodeJsonString(templateData.dataAsSerializedJson)
	if err != nil {
		return stacktrace.Propagate(err, "There was an error decoding the data as a JSON string: '%s'", templateData.dataAsSerializedJson)
	}

	// Create all parent directories to account for nesting
	destinationFileDir := path.Dir(destinationAbsoluteFilePath)
	if err := os.MkdirAll(destinationFileDir, folderPermissionForRenderedTemplates); err != nil {
		return stacktrace.Propagate(err, "There was an error in creating the parent directory '%s' to write the file '%s' into.", destinationFileDir, destinationAbsoluteFilePath)
	}

	renderedTemplateFile, err := os.Create(destinationAbsoluteFilePath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while creating temporary file to render template into for file '%s'.", destinationAbsoluteFilePath)
	}
	defer renderedTemplateFile.Close()

	if err = templateData.template.Execute(renderedTemplateFile, decodedData); err != nil {
		return stacktrace.Propagate(err, "An error occurred while writing the rendered template to destination '%s'", destinationAbsoluteFilePath)
	}
	return nil
}

func decodeJsonString(dataAsSerializedJson string) (interface{}, error) {
	dataAsSerializedJsonBytes := []byte(dataAsSerializedJson)
	dataJsonReader := bytes.NewReader(dataAsSerializedJsonBytes)

	// We don't use standard json.Unmarshal as that converts large integers to floats
	// Using this custom decoder we get the json.Number representation which is closer to other json implementations
	// This talks about the issue further https://github.com/square/go-jose/issues/351#issuecomment-847193900
	decoder := json.NewDecoder(dataJsonReader)
	decoder.UseNumber()

	var decodedData interface{}
	if err := decoder.Decode(&decodedData); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while decoding the template data json '%s'", dataAsSerializedJson)
	}
	return decodedData, nil
}

func generateUniqueTemplateName(templateString string) (string, error) {
	md5Hash := md5.New()
	if _, err := md5Hash.Write([]byte(templateString)); err != nil {
		return "", stacktrace.Propagate(err, "Unable to hash content of template '%s'", templateString)
	}
	templateStringHash := md5Hash.Sum(nil)
	templateName := fmt.Sprintf("%s%s", templateNamePrefix, templateStringHash)
	return templateName, nil
}
