/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclaves

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
type TemplateAndData struct {
	Template     string
	TemplateData interface{}
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func NewTemplateAndData(template string, templateData interface{}) *TemplateAndData {
	return &TemplateAndData{template, templateData}
}
