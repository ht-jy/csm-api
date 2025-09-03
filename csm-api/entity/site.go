package entity

import (
	"github.com/guregu/null"
)

/**
 * @author 작성자: 김진우
 * @created 작성일: 2025-02-12
 * @modified 최종 수정일:
 * @modifiedBy 최종 수정자:
 * @modified description
 * -
 */

// struct: 현장 관리 응답 구조체
type SiteRes struct {
	Site Sites `json:"site"`
	Code Codes `json:"code"`
}

type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Site struct {
	RowNum             null.Int    `json:"rnum" db:"RNUM"`
	Sno                null.Int    `json:"sno" db:"SNO"`
	SiteNm             null.String `json:"site_nm" db:"SITE_NM"`
	WorkRate           null.Float  `json:"work_rate" db:"WORK_RATE"`
	Etc                null.String `json:"etc" db:"ETC"`
	LocCode            null.String `json:"loc_code" db:"LOC_CODE"`
	LocName            null.String `json:"loc_name" db:"LOC_NAME"`
	Status             null.String `json:"status" db:"STATUS"`
	DefaultJno         null.Int    `json:"default_jno" db:"DEFAULT_JNO"`
	DefaultProjectName null.String `json:"default_project_name" db:"DEFAULT_PROJECT_NAME"`
	DefaultProjectNo   null.String `json:"default_project_no" db:"DEFAULT_PROJECT_NO"`
	CurrentSiteStats   null.String `json:"current_site_stats" db:"CURRENT_SITE_STATS"`
	SelectDate         null.String `json:"select_date" db:"SELECT_DATE"`
	Base

	ProjectList *ProjectInfos       `json:"project_list"`
	SitePos     *SitePos            `json:"site_pos"`
	SiteDate    *SiteDate           `json:"site_date"`
	Weather     WeatherSrtEntityRes `json:"whether"`
}

// struct: 현장 데이터 json 배열 구조체
type Sites []*Site

type ReqSite struct {
	Sno null.Int `json:"sno" db:"SNO"`
	Jno null.Int `json:"jno" db:"JNO"`
	Base
}

// 공정률 저장
type SiteWorkRate struct {
	Sno        null.Int    `json:"sno" db:"SNO"`
	Jno        null.Int    `json:"jno" db:"JNO"`
	RecordDate null.Time   `json:"record_date" db:"RECORD_DATE"`
	WorkRate   null.Float  `json:"work_rate" db:"WORK_RATE"`
	IsWorkRate null.String `json:"is_work_rate" db:"IS_WORK_RATE"`
	SearchDate null.String `json:"search_date" db:"SEARCH_DATE"`
	Base
}

type SiteWorkRates []*SiteWorkRate
