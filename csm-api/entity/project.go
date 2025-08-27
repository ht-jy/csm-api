package entity

import (
	"github.com/guregu/null"
)

type ProjectInfo struct {
	Sno                   null.Int    `json:"sno" db:"SNO"`
	Jno                   null.Int    `json:"jno" db:"JNO"`
	Status                null.String `json:"status" db:"STATUS"`
	IsDefault             null.String `json:"is_default" db:"IS_DEFAULT"`
	ProjectType           null.String `json:"project_type" db:"PROJECT_TYPE"`
	ProjectTypeNm         null.String `json:"project_type_nm" db:"PROJECT_TYPE_NM"`
	ProjectNo             null.String `json:"project_no" db:"PROJECT_NO"`
	ProjectNm             null.String `json:"project_nm" db:"PROJECT_NM"`
	ProjectYear           null.Int    `json:"project_year" db:"PROJECT_YEAR"`
	ProjectLoc            null.String `json:"project_loc" db:"PROJECT_LOC"`
	ProjectCode           null.String `json:"project_code" db:"PROJECT_CODE"`
	ProjectCodeName       null.String `json:"project_code_name" db:"PROJECT_CODE_NAME"`
	SiteNm                null.String `json:"site_nm" db:"SITE_NM"`
	CompCode              null.String `json:"comp_code" db:"COMP_CODE"`
	CompNick              null.String `json:"comp_nick" db:"COMP_NICK"`
	CompName              null.String `json:"comp_name" db:"COMP_NAME"`
	CompEtc               null.String `json:"comp_etc" db:"COMP_ETC"`
	OrderCompCode         null.String `json:"order_comp_code" db:"ORDER_COMP_CODE"`
	OrderCompNick         null.String `json:"order_comp_nick" db:"ORDER_COMP_NICK"`
	OrderCompName         null.String `json:"order_comp_name" db:"ORDER_COMP_NAME"`
	OrderCompJobName      null.String `json:"order_comp_job_name" db:"ORDER_COMP_JOB_NAME"`
	ProjectLocName        null.String `json:"project_loc_name" db:"PROJECT_LOC_NAME"`
	JobPm                 null.String `json:"job_pm" db:"JOB_PM"`
	JobPmNm               null.String `json:"job_pm_nm" db:"JOB_PM_NAME"`
	JobPe                 null.String `json:"job_pe" db:"JOB_PE"`
	ProjectStdt           null.Time   `json:"project_stdt" db:"PROJECT_STDT"`
	ProjectEddt           null.Time   `json:"project_eddt" db:"PROJECT_EDDT"`
	ProjectRegDate        null.Time   `json:"project_reg_date" db:"PROJECT_REG_DATE"`
	ProjectModDate        null.Time   `json:"project_mod_date" db:"PROJECT_MOD_DATE"`
	ProjectState          null.String `json:"project_state" db:"PROJECT_STATE"`
	ProjectStateNm        null.String `json:"project_state_nm" db:"PROJECT_STATE_NM"`
	MocNo                 null.String `json:"moc_no" db:"MOC_NO"`
	WoNo                  null.String `json:"wo_no" db:"WO_NO"`
	WorkerCountAll        null.Int    `json:"worker_count_all" db:"WORKER_COUNT_ALL"`
	WorkerCountDate       null.Int    `json:"worker_count_date" db:"WORKER_COUNT_DATE"`
	WorkerCountHtenc      null.Int    `json:"worker_count_htenc" db:"WORKER_COUNT_HTENC"`
	WorkerCountWork       null.Int    `json:"worker_count_work" db:"WORKER_COUNT_WORK"`
	WorkerCountSafe       null.Int    `json:"worker_count_safe" db:"WORKER_COUNT_SAFE"`
	WorkerCountManager    null.Int    `json:"worker_count_manager" db:"WORKER_COUNT_MANAGER"`
	WorkerCountNotManager null.Int    `json:"worker_count_not_manager" db:"WORKER_COUNT_NOT_MANAGER"`
	EquipCount            null.Int    `json:"equip_count" db:"EQUIP_COUNT"`
	WorkRate              null.Int    `json:"work_rate" db:"WORK_RATE"`
	IsWorkRate            null.String `json:"is_work_rate" db:"IS_WORK_RATE"`
	CancelDay             null.Int    `json:"cancel_day" db:"CANCEL_DAY"`
	Base

	ProjectPeList    *UserPeInfos   `json:"project_pe_list"`
	DailyContentList *ProjectDailys `json:"daily_content_list"`
}

type ProjectInfos []*ProjectInfo

type ProjectSafeCount struct {
	Sno       null.Int `json:"sno" db:"SNO"`
	Jno       null.Int `json:"jno" db:"JNO"`
	SafeCount null.Int `json:"safe_count" db:"SAFE_COUNT"`
}
type ProjectSafeCounts []*ProjectSafeCount

// nonUsedProject
type NonUsedProject struct {
	Rnum          null.Int    `json:"rnum" db:"RNUM"`
	Jno           null.Int    `json:"jno" db:"JNO"`
	JobNo         null.String `json:"job_no" db:"JOB_NO"`
	JobName       null.String `json:"job_name" db:"JOB_NAME"`
	JobYear       null.Int    `json:"job_year" db:"JOB_YEAR"`
	JobSd         null.String `json:"job_sd" db:"JOB_SD"`
	JobEd         null.String `json:"job_ed" db:"JOB_ED"`
	JobPmNm       null.String `json:"job_pm_nm" db:"JOB_PM_NM"`
	DutyName      null.String `json:"duty_name" db:"DUTY_NAME"`
	CompName      null.String `json:"comp_name" db:"COMP_NAME"`
	OrderCompName null.String `json:"order_comp_name" db:"ORDER_COMP_NAME"`
	CdNm          null.String `json:"cd_nm" db:"CD_NM"`
}

type NonUsedProjects []*NonUsedProject

// 프로젝트(IRIS_SITE_JOB) 추가/수정 구조체
type ReqProject struct {
	Sno       null.Int    `json:"sno" db:"SNO"`
	Jno       null.Int    `json:"jno" db:"JNO"`
	IsUsed    null.String `json:"is_used" db:"IS_USED"`
	IsDefault null.String `json:"is_default" db:"IS_DEFAULT"`
	WorkRate  null.Int    `json:"work_rate" db:"WORK_RATE"`
	Base
}
