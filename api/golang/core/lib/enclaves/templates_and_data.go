/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclaves

// TemplateAndData Docs available at https://docs.kurtosis.com/sdk#templateanddata
type TemplateAndData struct {
	template     string
	templateData interface{}
}

// NewTemplateAndData Docs available at https://docs.kurtosis.com/sdk#templateanddata
func NewTemplateAndData(template string, templateData interface{}) *TemplateAndData {
	return &TemplateAndData{template, templateData}
}
