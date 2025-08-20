package store

import (
	"context"
	"csm-api/entity"
	"csm-api/utils"
	"database/sql"
	"errors"
	"strconv"
)

// TBM 엑셀 차수 조회
func (r *Repository) GetTbmOrder(ctx context.Context, db Queryer, tbm entity.Tbm) (string, error) {
	var order sql.NullInt64

	query := `
		SELECT 
			MAX(TBM_ORDER) 
		FROM IRIS_TBM_SET
		WHERE SNO = :1
		AND DEPARTMENT = :2
		AND TRUNC(TBM_DATE) = TRUNC(:3)`

	if err := db.GetContext(ctx, &order, query, tbm.Sno, tbm.Department, tbm.TbmDate); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "0", nil
		}
		return "0", utils.CustomErrorf(err)
	}

	if !order.Valid {
		return "0", nil
	}

	return strconv.FormatInt(order.Int64+1, 10), nil
}

// TBM 엑셀 정보 저장
func (r *Repository) AddTbmExcel(ctx context.Context, tx Execer, tbms []entity.Tbm) error {
	agent := utils.GetAgent()

	query := `
		INSERT INTO IRIS_TBM_SET(SNO, DEPARTMENT, DISC_NAME, USER_NM, TBM_DATE, TBM_ORDER, REG_DATE, REG_USER, REG_UNO, REG_AGENT)
		VALUES(:1, :2, :3, :4, :5, :6, SYSDATE, :7, :8, :9)`

	for _, tbm := range tbms {
		if _, err := tx.ExecContext(ctx, query, tbm.Sno, tbm.Department, tbm.DiscName, tbm.UserNm, tbm.TbmDate, tbm.TbmOrder, tbm.RegUser, tbm.RegUno, agent); err != nil {
			return utils.CustomErrorf(err)
		}
	}
	return nil
}

// 퇴직공제 현장명 조회
func (r *Repository) GetDeductionJobNameByJno(ctx context.Context, db Queryer, jno int64) (string, error) {
	var name sql.NullString

	query := `SELECT JOB_NAME FROM S_JOB_INFO WHERE JNO = :1`

	if err := db.GetContext(ctx, &name, query, jno); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", utils.CustomErrorf(err)
	}
	if !name.Valid {
		return "", nil
	}
	return name.String, nil
}

// 퇴직공제 차수 조회
func (r *Repository) GetDeductionOrder(ctx context.Context, db Queryer, tbm entity.Deduction) (string, error) {
	var order sql.NullInt64

	query := `
		SELECT 
			MAX(DEDUCT_ORDER) 
		FROM IRIS_DEDUCTION_SET
		WHERE SNO = :1
		AND TRUNC(RECORD_DATE) = TRUNC(:2)`

	if err := db.GetContext(ctx, &order, query, tbm.Sno, tbm.RecordDate); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "0", nil
		}
		return "0", utils.CustomErrorf(err)
	}
	if !order.Valid {
		return "0", nil
	}
	return strconv.FormatInt(order.Int64+1, 10), nil
}

// 퇴직공제 엑셀 정보 저장
func (r *Repository) AddDeductionExcel(ctx context.Context, tx Execer, tbms []entity.Deduction) error {
	agent := utils.GetAgent()

	query := `
		INSERT INTO IRIS_DEDUCTION_SET(SNO, JNO, USER_NM, DEPARTMENT, GENDER, REG_NO, PHONE, IN_RECOG_TIME, OUT_RECOG_TIME, RECORD_DATE, DEDUCT_ORDER, REG_DATE, REG_USER, REG_UNO, REG_AGENT)
		VALUES(:1, :2, :3, :4, :5, :6, :7, :8, :9, :10, :11, SYSDATE, :12, :13, :14)`

	for _, tbm := range tbms {
		if _, err := tx.ExecContext(ctx, query, tbm.Sno, tbm.Jno, tbm.UserNm, tbm.Department, tbm.Gender, tbm.RegNo, tbm.Phone, tbm.InRecogTime, tbm.OutRecogTime, tbm.RecordDate, tbm.DeductOrder, tbm.RegUser, tbm.RegUno, agent); err != nil {
			return utils.CustomErrorf(err)
		}
	}
	return nil
}
