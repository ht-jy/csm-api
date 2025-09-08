package entity

import (
	"github.com/guregu/null"
)

/**
 * @author 작성자: 김진우
 * @created 작성일: 2025-02-18
 * @modified 최종 수정일: 2025-02-26
 * @modifiedBy 최종 수정자: 정지영
 * @modified description
 * - 현장소장 및 안전관리자 userId, UserInfo 추가
 */

// struct: Begin:job 정보
type JobInfo struct {
	RowNum        null.Int    `json:"rnum" db:"RNUM"`
	Jno           null.Int    `json:"jno" db:"JNO"`
	Sno           null.Int    `json:"sno" db:"SNO"`
	JobName       null.String `json:"job_name" db:"JOB_NAME"`
	JobNo         null.String `json:"job_no" db:"JOB_NO"`
	JobSd         null.String `json:"job_sd" db:"JOB_SD"`
	JobEd         null.String `json:"job_ed" db:"JOB_ED"`
	CompName      null.String `json:"comp_name" db:"COMP_NAME"`
	OrderCompName null.String `json:"order_comp_name" db:"ORDER_COMP_NAME"`
	JobPmName     null.String `json:"job_pm_name" db:"JOB_PM_NAME"`
	JobPmDutyName null.String `json:"job_pm_duty_name" db:"JOB_PM_DUTY_NAME"`
	CdNm          null.String `json:"cd_nm" db:"CD_NM"`
}
type JobInfos []*JobInfo

// struct: End:job 정보

// struct: Begin::현장소장|안전관리자
type Manager struct {
	Uno        null.Int    `json:"uno" db:"UNO"`
	Jno        null.Int    `json:"jno" db:"JNO"`
	UserName   null.String `json:"user_name" db:"USER_NAME"`
	DutyName   null.String `json:"duty_name" db:"DUTY_NAME"`
	UserId     null.String `json:"user_id" db:"USER_ID"`
	TeamLeader null.String `json:"team_leader" db:"TEAM_LEADER"`
	UserInfo   null.String `json:"user_info" db:"USER_INFO"`
}
type Managers []*Manager

// struct: end::현장소장|안전관리자

// struct: Begin::관리감독자
type Supervisor struct {
	Uno               null.Int    `json:"uno" db:"UNO"`
	Jno               null.Int    `json:"jno" db:"JNO"`
	UserName          null.String `json:"user_name" db:"USER_NAME"`
	UserId            null.String `json:"user_id" db:"USER_ID"`
	DutyName          null.String `json:"duty_name" db:"DUTY_NAME"`
	DutyCd            null.String `json:"duty_cd" db:"DUTY_CD"`
	JobdutyId         null.String `json:"jobduty_id" db:"JOBDUTY_ID"`
	JoinDate          null.Time   `json:"join_date" db:"JOIN_DATE"`
	FuncNo            null.String `json:"func_no" db:"FUNC_NO"`
	CdNm              null.String `json:"cd_nm" db:"CD_NM"`
	SysSafe           null.String `json:"sys_safe" db:"SYS_SAFE"`
	IsSiteManager     null.String `json:"is_site_manager" db:"IS_SITE_MANAGER"`
	IsTempSiteManager null.String `json:"is_temp_site_manager" db:"IS_TEMP_SITE_MANAGER"`
}
type Supervisors []*Supervisor

// struct: end::관리감독자

// struct: Begin::협력업체 정보
type CompanyInfo struct {
	Jno       null.Int    `json:"jno" db:"JNO"`
	Cno       null.Int    `json:"cno" db:"CNO"`
	Id        null.String `json:"id" db:"ID"`
	Pw        null.String `json:"pw" db:"PW"`
	Cellphone null.String `json:"cellphone" db:"CELLPHONE"`
	Email     null.String `json:"email" db:"EMAIL"`
	UserName  null.String `json:"username" db:"USER_NAME"`
	DutyName  null.String `json:"duty_name" db:"DUTY_NAME"`
}
type CompanyInfos []*CompanyInfo

type CompanyInfoRes struct {
	Jno        int64   `json:"jno"`
	Cno        int64   `json:"cno"`
	CompNameKr string  `json:"comp_name_kr"`
	WorkerName string  `json:"worker_name"`
	Id         string  `json:"id"`
	Cellphone  string  `json:"cellphone"`
	Email      string  `json:"email"`
	WorkInfo   []int64 `json:"work_infos"`
}
type CompanyInfoResList []*CompanyInfoRes

// struct: End::협력업체 정보

// struct: Begin::공종 정보
type WorkInfo struct {
	Cno      null.Int    `json:"Cno" db:"CNO"`
	Jno      null.Int    `json:"jno" db:"JNO"`
	FuncNo   null.Int    `json:"func_no" db:"FUNC_NO"`
	FuncName null.String `json:"func_name" db:"FUNC_NAME"`
}
type WorkInfos []*WorkInfo

// struct: End::공종 정보

// struct: Begin::협력업체 리스트 api
type CompanyApiReq struct {
	ResultType string           `json:"ResultType"`
	ValueType  string           `json:"ValueType"`
	Value      CompanyApiValues `json:"Value"`
}

type CompanyApiValue struct {
	Cno          float64 `json:"CNO"`
	Jno          float64 `json:"JNO"`
	CompCno      float64 `json:"COMP_CNO"`
	ProjectName  string  `json:"PROJECT_NAME"`
	ItemDesc     string  `json:"ITEM_DESC"`
	ContractKind string  `json:"CONTRACT_KIND"`
	ContractDate string  `json:"CONTRACT_DATE"`
	DeliDate     string  `json:"DELI_DATE"`
	RegDate      string  `json:"REG_DATE"`
	ModDate      string  `json:"MOD_DATE"`
	CompNameKr   string  `json:"COMP_NAME_KR"`
	CompRegNo    string  `json:"COMP_REG_NO"`
	CompCeoName  string  `json:"COMP_CEO_NAME"`
	WorkerName   string  `json:"WORKER_NAME"`
}
type CompanyApiValues []*CompanyApiValue

// struct: End::협력업체 리스트 api
