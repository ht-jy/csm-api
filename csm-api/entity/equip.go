package entity

import "github.com/guregu/null"

type Equip struct {
	Sno        null.Int    `json:"sno" db:"SNO"`
	Jno        null.Int    `json:"jno" db:"JNO"`
	Cnt        null.Int    `json:"cnt" db:"CNT"`
	JobName    null.String `json:"job_name" db:"JOB_NAME"`
	RecordDate null.Time   `json:"record_date" db:"RECORD_DATE"`
}

type Equips []*Equip
