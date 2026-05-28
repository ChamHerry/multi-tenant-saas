// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/google/uuid"
)

// ProcessSteps is the golang structure for table process_steps.
type ProcessSteps struct {
	RepoId         uuid.UUID `json:"repo_id"          orm:"repo_id"          description:""` //
	ProcessId      string    `json:"process_id"       orm:"process_id"       description:""` //
	StepIndex      int       `json:"step_index"       orm:"step_index"       description:""` //
	SymbolUid      string    `json:"symbol_uid"       orm:"symbol_uid"       description:""` //
	RelationToNext string    `json:"relation_to_next" orm:"relation_to_next" description:""` //
}
