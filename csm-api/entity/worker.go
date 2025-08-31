package entity

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/guregu/null"
	"io"
)

type Worker struct {
	RowNum      null.Int    `json:"rnum" db:"RNUM"`
	IrisNo      null.Int    `json:"iris_no" db:"IRIS_NO"`
	UserKey     null.String `json:"user_key" db:"USER_KEY"`
	Sno         null.Int    `json:"sno" db:"SNO"` //현장 고유번호
	SiteNm      null.String `json:"site_nm" db:"SITE_NM"`
	Jno         null.Int    `json:"jno" db:"JNO"` //프로젝트 고유번호
	JobName     null.String `json:"job_name" db:"JOB_NAME"`
	UserId      null.String `json:"user_id" db:"USER_ID"` //근로자 아이디
	AfterUserId null.String `json:"after_user_id" db:"AFTER_USER_ID"`
	UserNm      null.String `json:"user_nm" db:"USER_NM"`       //근로자명
	Department  null.String `json:"department" db:"DEPARTMENT"` //부서or조직
	DiscName    null.String `json:"disc_name" db:"DISC_NAME"`   // 공종명
	Phone       null.String `json:"phone" db:"PHONE"`
	WorkerType  null.String `json:"worker_type" db:"WORKER_TYPE"`
	CodeNm      null.String `json:"code_nm" db:"CODE_NM"`
	IsUse       null.String `json:"is_use" db:"IS_USE"`
	IsRetire    null.String `json:"is_retire" db:"IS_RETIRE"`
	IsManage    null.String `json:"is_manage" db:"IS_MANAGE"`
	RetireDate  null.Time   `json:"retire_date" db:"RETIRE_DATE"`
	RecordDate  null.String `json:"record_date" db:"RECORD_DATE"`
	RegNo       null.String `json:"reg_no" db:"REG_NO"`
	FailReason  null.String `json:"fail_reason" db:"FAIL_REASON"`
	WorkerReason
	Base
}
type Workers []*Worker

type WorkerDaily struct {
	RowNum          null.Int    `json:"rnum" db:"RNUM"`
	IrisNo          null.Int    `json:"iris_no" db:"IRIS_NO"`
	Dno             null.String `json:"dno" db:"DNO"`
	Sno             null.Int    `json:"sno" db:"SNO"` //현장 고유번호
	Jno             null.Int    `json:"jno" db:"JNO"` //프로젝트 고유번호
	JobName         null.String `json:"job_name" db:"JOB_NAME"`
	UserKey         null.String `json:"user_key" db:"USER_KEY"`
	UserId          null.String `json:"user_id" db:"USER_ID"` //근로자 아이디
	UserNm          null.String `json:"user_nm" db:"USER_NM"`
	Department      null.String `json:"department" db:"DEPARTMENT"`
	DiscName        null.String `json:"disc_name" db:"DISC_NAME"` // 공종명
	Phone           null.String `json:"phone" db:"PHONE"`
	RegNo           null.String `json:"reg_no" db:"REG_NO"`
	RecordDate      null.Time   `json:"record_date" db:"RECORD_DATE"`
	InRecogTime     null.Time   `json:"in_recog_time" db:"IN_RECOG_TIME"`   //출근시간
	OutRecogTime    null.Time   `json:"out_recog_time" db:"OUT_RECOG_TIME"` //퇴근시간
	WorkState       null.String `json:"work_state" db:"WORK_STATE"`
	IsDeadline      null.String `json:"is_deadline" db:"IS_DEADLINE"`
	IsOvertime      null.String `json:"is_overtime" db:"IS_OVERTIME"`
	CompareState    null.String `json:"compare_state" db:"COMPARE_STATE"`
	WorkHour        null.Float  `json:"work_hour" db:"WORK_HOUR"`
	DeviceNm        null.String `json:"device_nm" db:"DEVICE_NM"`
	SearchStartTime null.String `json:"search_start_time" db:"SEARCH_START_TIME"`
	SearchEndTime   null.String `json:"search_end_time" db:"SEARCH_END_TIME"`
	AfterJno        null.Int    `json:"after_jno" db:"AFTER_JNO"`
	BeforeState     null.String `json:"before_state" db:"BEFORE_STATE"`
	AfterState      null.String `json:"after_state" db:"AFTER_STATE"`
	Message         null.String `json:"message" db:"MESSAGE"`
	FailReason      null.String `json:"fail_reason" db:"FAIL_REASON"`
	WorkerReason
	Base
}
type WorkerDailys []*WorkerDaily

type WorkerOverTime struct {
	BeforeCno    null.Int  `json:"before_cno" db:"BEFORE_CNO"`         // 출근한 날 CNO
	AfterCno     null.Int  `json:"after_cno" db:"AFTER_CNO"`           // 퇴근한 날 CNO
	OutRecogTime null.Time `json:"out_recog_time" db:"OUT_RECOG_TIME"` // 퇴근시간
}
type WorkerOverTimes []*WorkerOverTime

type WorkerDailyExcel struct {
	RegNo      string
	Department string
	UserNm     string
	Phone      string
	WorkDate   string
	InTime     string
	OutTime    string
	WorkHour   string
}

type RecordDailyWorkerReq struct {
	Jno       null.Int    `json:"jno" db:"JNO"`
	StartDate null.String `json:"start_date" db:"START_DATE"`
	EndDate   null.String `json:"end_date" db:"END_DATE"`
}

type RecordDailyWorkerRes struct {
	JobName      null.String `json:"job_name" db:"JOB_NAME"`
	UserNm       null.String `json:"user_nm" db:"USER_NM"`
	Department   null.String `json:"department" db:"DEPARTMENT"`
	Phone        null.String `json:"phone" db:"PHONE"`
	RecordDate   null.Time   `json:"record_date" db:"RECORD_DATE"`
	InRecogTime  null.Time   `json:"in_recog_time" db:"IN_RECOG_TIME"`
	OutRecogTime null.Time   `json:"out_recog_time" db:"OUT_RECOG_TIME"`
	WorkHour     null.Float  `json:"work_hour" db:"WORK_HOUR"`
	IsDeadline   null.String `json:"is_deadline" db:"IS_DEADLINE"`
}

type DailyWorkerExcel struct {
	StartDate   string        `json:"start_date" db:"START_DATE"`
	EndDate     string        `json:"end_date" db:"END_DATE"`
	WorkerExcel []WorkerExcel `json:"worker_excel"`
}

type WorkerExcel struct {
	JobName         string            `json:"job_name" db:"JOB_NAME"`
	UserNm          string            `json:"user_nm" db:"USER_NM"`
	Department      string            `json:"department" db:"DEPARTMENT"`
	Phone           string            `json:"phone" db:"PHONE"`
	SumWorkHour     float64           `json:"sum_work_hour" db:"SUM_WORK_HOUR"`
	SumWorkDate     int64             `json:"sum_work_date" db:"SUM_WORK_DATE"`
	WorkerTimeExcel []WorkerTimeExcel `json:"worker_time_excel"`
}

type WorkerTimeExcel struct {
	RecordDate   string  `json:"record_date" db:"RECORD_DATE"`
	InRecogTime  string  `json:"in_recog_time" db:"IN_RECOG_TIME"`
	OutRecogTime string  `json:"out_recog_time" db:"OUT_RECOG_TIME"`
	WorkHour     float64 `json:"work_hour" db:"WORK_HOUR"`
	IsDeadline   string  `json:"is_deadline" db:"IS_DEADLINE"`
}

type WorkerReason struct {
	Cno        null.Int    `json:"cno" db:"CNO"`
	Reason     null.String `json:"reason" db:"REASON"`
	ReasonType null.String `json:"reason_type" db:"REASON_TYPE"`
	HisStatus  null.String `json:"his_status" db:"HIS_STATUS"`
	HisName    null.String `json:"his_name" db:"HIS_NAME"`
}

func (c *Worker) Decode(encKey string, secretKey string) (string, error) {
	key := []byte(secretKey)
	ct, err := base64.StdEncoding.DecodeString(encKey)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ct) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, data := ct[:nonceSize], ct[nonceSize:]

	plain, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func (c *Worker) Encode(regNo string, secretKey string) (string, error) {
	key := []byte(secretKey)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(regNo), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
