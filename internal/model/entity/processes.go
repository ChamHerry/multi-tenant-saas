// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// Processes is the golang structure for table processes.
type Processes struct {
	RepoId        uuid.UUID   `json:"repo_id"         orm:"repo_id"         description:""` //
	Id            string      `json:"id"              orm:"id"              description:""` //
	Label         string      `json:"label"           orm:"label"           description:""` //
	ProcessType   string      `json:"process_type"    orm:"process_type"    description:""` //
	StepCount     int         `json:"step_count"      orm:"step_count"      description:""` //
	Communities   []string    `json:"communities"     orm:"communities"     description:""` //
	EntryPointUid string      `json:"entry_point_uid" orm:"entry_point_uid" description:""` //
	TerminalUid   string      `json:"terminal_uid"    orm:"terminal_uid"    description:""` //
	Summary       string      `json:"summary"         orm:"summary"         description:""` //
	Metadata      *gjson.Json `json:"metadata"        orm:"metadata"        description:""` //
	CreatedAt     time.Time   `json:"created_at"      orm:"created_at"      description:""` //
	UpdatedAt     time.Time   `json:"updated_at"      orm:"updated_at"      description:""` //
}
