package models

import (
	model "github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
)

// InitMessage ...
type InitMessage struct {
	Type string                 `json:"type"`
	Msg  model.BitriseDataModel `json:"msg"`
}

//Message ...
type Message struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

//SaveMessage ...
type SaveMessage struct {
	Type string                 `json:"type"`
	Msg  model.BitriseDataModel `json:"msg"`
}

//Workflows ...
type Workflows struct {
	FormatVersion string                 `json:"format_version" yaml:"format_version"`
	Workflows     map[string]interface{} `json:"workflows" yaml:"workflows"`
}
