/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclaves

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
type TemplateAndData struct {
	Template     string
	TemplateData any
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func NewTemplateAndData(template string, templateData any) *TemplateAndData {
	return &TemplateAndData{template, templateData}
}