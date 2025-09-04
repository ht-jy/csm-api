package store

import (
	"context"
	"csm-api/entity"
	"csm-api/utils"
	"fmt"
)

// 일일 근로자 비교 - 근로자 리스트
func (r *Repository) GetDailyWorkerList(ctx context.Context, db Queryer, compare entity.Compare, retry string, order string) (entity.WorkerDailys, error) {
	var list entity.WorkerDailys

	var columns []string
	columns = append(columns, "T2.USER_NM")
	columns = append(columns, "T2.DEPARTMENT")
	columns = append(columns, "T2.USER_ID")
	columns = append(columns, "T3.DEVICE_NM")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	var orderBy string
	if order == "" {
		orderBy = `
			ORDER BY
				CASE COMPARE_STATE
					WHEN 'S' THEN 1
					WHEN 'X' THEN 2		
					WHEN 'W' THEN 3
					WHEN 'C' THEN 4
					ELSE 5
				END,
				CASE 
					WHEN IN_RECOG_TIME IS NULL THEN OUT_RECOG_TIME 
					WHEN OUT_RECOG_TIME IS NULL THEN IN_RECOG_TIME 
					ELSE GREATEST(IN_RECOG_TIME, OUT_RECOG_TIME) 
				END
				DESC NULLS LAST`
	} else {
		orderBy = fmt.Sprintf(`ORDER BY %s`, order)
	}

	query := fmt.Sprintf(`
		SELECT
		    T1.USER_KEY,
			T1.SNO,
			T1.JNO,
			T2.USER_ID,
			T2.USER_NM,
			T2.PHONE,
			COMMON.FUNC_DECODE(T2.REG_NO) AS REG_NO,
			CASE
				WHEN INSTR(T2.DEPARTMENT, ' ', -1) > 0 THEN SUBSTR(T2.DEPARTMENT, 1, INSTR(T2.DEPARTMENT, ' ', -1) - 1)
				ELSE T2.DEPARTMENT
			END AS DEPARTMENT,
			T2.DISC_NAME,
			T1.IN_RECOG_TIME,
			T1.OUT_RECOG_TIME,
			TRUNC(T1.RECORD_DATE) AS RECORD_DATE,
			T1.COMPARE_STATE,
			T1.IS_DEADLINE,
			T3.DEVICE_NM
		FROM IRIS_WORKER_DAILY_SET T1
		LEFT JOIN IRIS_WORKER_SET T2 ON T1.SNO = T2.SNO AND T1.USER_KEY = T2.USER_KEY --T1.SNO = T2.SNO AND T1.USER_ID = T2.USER_ID
		LEFT JOIN IRIS_DEVICE_SET T3 ON T1.DNO = T3.DNO 
		WHERE TRUNC(T1.RECORD_DATE) = TRUNC(:1)
			AND T1.SNO = :2
			AND T2.IS_DEL = 'N'
			AND (
				T1.JNO = :3 
				OR (T1.JNO != :4 AND T1.COMPARE_STATE NOT IN ('S', 'X'))
			)
			%s
			%s`, retryCondition, orderBy)

	if err := db.SelectContext(ctx, &list, query, compare.RecordDate, compare.Sno, compare.Jno, compare.Jno); err != nil {
		return list, utils.CustomErrorf(err)
	}
	return list, nil
}

// 일일 근로자 비교 - TBM 리스트
func (r *Repository) GetTbmList(ctx context.Context, db Queryer, compare entity.Compare, retry string, order string) ([]entity.Tbm, error) {
	var list []entity.Tbm

	var columns []string
	columns = append(columns, "A.USER_NM")
	columns = append(columns, "A.DEPARTMENT")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	var orderBy string
	if order == "" {
		orderBy = `ORDER BY TBM_DATE DESC NULLS LAST`
	} else {
		orderBy = fmt.Sprintf(`ORDER BY %s`, order)
	}

	query := fmt.Sprintf(`
		SELECT 
			A.SNO,
			A.JNO,
			A.DEPARTMENT,
			A.DISC_NAME,
			A.USER_NM,
			TRUNC(A.TBM_DATE) AS TBM_DATE,
			A.TBM_ORDER
		FROM IRIS_TBM_SET A
		JOIN (
			SELECT 
				SNO, JNO, DEPARTMENT, USER_NM, MAX(TBM_ORDER) AS MAX_ORDER
			FROM IRIS_TBM_SET
			WHERE SNO = :1
			  AND TRUNC(TBM_DATE) = TRUNC(:2)
			GROUP BY SNO, JNO, DEPARTMENT, USER_NM
		) B
		  ON A.SNO = B.SNO
		 AND A.DEPARTMENT = B.DEPARTMENT
		 AND A.USER_NM = B.USER_NM
		 AND A.TBM_ORDER = B.MAX_ORDER
		WHERE A.SNO = :3
		AND TRUNC(A.TBM_DATE) = TRUNC(:4)
		AND (
			A.JNO IS NULL 
			OR A.JNO = :5
		)
		%s
		%s`, retryCondition, orderBy)

	if err := db.SelectContext(ctx, &list, query, compare.Sno, compare.RecordDate, compare.Sno, compare.RecordDate, compare.Jno); err != nil {
		return list, utils.CustomErrorf(err)
	}
	return list, nil
}

// 일일 근로자 비교 - 퇴직공제 리스트
func (r *Repository) GetDeductionList(ctx context.Context, db Queryer, compare entity.Compare, retry string, order string) ([]entity.Deduction, error) {
	var list []entity.Deduction

	var columns []string
	columns = append(columns, "A.USER_NM")
	columns = append(columns, "A.DEPARTMENT")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	var orderBy string
	if order == "" {
		orderBy = `
			ORDER BY (
				CASE 
					WHEN IN_RECOG_TIME IS NULL THEN OUT_RECOG_TIME 
					WHEN OUT_RECOG_TIME IS NULL THEN IN_RECOG_TIME 
					ELSE GREATEST(IN_RECOG_TIME, OUT_RECOG_TIME) 
				END
			) DESC NULLS LAST`
	} else {
		orderBy = fmt.Sprintf(`ORDER BY %s`, order)
	}

	query := fmt.Sprintf(`
		SELECT 
			A.SNO,
			A.JNO,
			A.USER_NM,
			A.GENDER,
			A.PHONE,
			A.REG_NO,
			A.DEPARTMENT,
			A.IN_RECOG_TIME,
			A.OUT_RECOG_TIME,
			TRUNC(A.RECORD_DATE) AS RECORD_DATE,
			A.DEDUCT_ORDER
		FROM IRIS_DEDUCTION_SET A
		JOIN (
			SELECT 
				SNO, JNO, USER_NM, REG_NO, DEPARTMENT, GENDER, MAX(DEDUCT_ORDER) AS MAX_ORDER
			FROM IRIS_DEDUCTION_SET
			WHERE SNO = :1
			  AND TRUNC(RECORD_DATE) = TRUNC(:2)
			GROUP BY SNO, JNO, USER_NM, REG_NO, DEPARTMENT, GENDER
		) B
		  ON A.SNO = B.SNO
		 AND A.USER_NM = B.USER_NM
		 AND A.REG_NO = B.REG_NO
		 AND A.DEPARTMENT = B.DEPARTMENT
		 AND A.GENDER = B.GENDER
		 AND A.DEDUCT_ORDER = B.MAX_ORDER
		WHERE A.SNO = :3
		 AND TRUNC(A.RECORD_DATE) = TRUNC(:4)
		 AND (
			A.JNO IS NULL
			OR A.JNO = :5
		 )
		%s
		%s`, retryCondition, orderBy)

	if err := db.SelectContext(ctx, &list, query, compare.Sno, compare.RecordDate, compare.Sno, compare.RecordDate, compare.Jno); err != nil {
		return list, utils.CustomErrorf(err)
	}
	return list, nil
}

// 근로자 비교 반영 - 근로자 정보: IRIS_WORKER_SET
// 선택한 프로젝트로 수정
func (r *Repository) ModifyWorkerCompareApply(ctx context.Context, tx Execer, workers entity.WorkerDailys) error {
	agent := utils.GetAgent()

	query := `
		UPDATE IRIS_WORKER_SET
		SET
			JNO = :1,
			MOD_DATE = SYSDATE,
			MOD_USER = :2,
			MOD_UNO = :3,
			MOD_AGENT = :4
		WHERE SNO = :5
		AND USER_KEY = :6
		AND IS_DEL = 'N'`

	for _, worker := range workers {
		if _, err := tx.ExecContext(ctx, query, worker.Jno, worker.RegUser, worker.RegUno, agent, worker.Sno, worker.UserKey); err != nil {
			return utils.CustomErrorf(err)
		}
	}
	return nil
}

// 근로자 비교 반영 - 근로자 일일 정보: IRIS_WORKER_DAILY_SET
// 반영상태, 선택한 프로젝트로 수정
func (r *Repository) ModifyDailyWorkerCompareApply(ctx context.Context, tx Execer, workers entity.WorkerDailys) error {
	agent := utils.GetAgent()

	query := `
		UPDATE IRIS_WORKER_DAILY_SET
		SET
			JNO = :1,
			COMPARE_STATE = :2,
			MOD_DATE = SYSDATE,
			MOD_USER = :3,
			MOD_UNO = :4,
			MOD_AGENT = :5
		WHERE SNO = :6
		AND USER_KEY = :7
		AND TRUNC(RECORD_DATE) = TRUNC(:8)`

	for _, worker := range workers {
		if _, err := tx.ExecContext(ctx, query, worker.Jno, worker.AfterState, worker.RegUser, worker.RegUno, agent, worker.Sno, worker.UserKey, worker.RecordDate); err != nil {
			return utils.CustomErrorf(err)
		}
	}
	return nil
}

// 근로자 비교 반영 - TBM 등록 정보: IRIS_TBM_SET
// 선택한 프로젝트로 수정
func (r *Repository) ModifyTbmCompareApply(ctx context.Context, tx Execer, workers entity.WorkerDailys) error {
	agent := utils.GetAgent()

	query := `
		UPDATE IRIS_TBM_SET
		SET
			JNO = :1,
			MOD_DATE = SYSDATE,
			MOD_USER = :2,
			MOD_UNO = :3,
			MOD_AGENT = :4
		WHERE ROWID = (
			SELECT ROWID FROM (
				SELECT ROWID
				FROM IRIS_TBM_SET
				WHERE SNO = :5
				AND USER_NM = :6
				AND DEPARTMENT = :7
				AND TRUNC(TBM_DATE) = TRUNC(:8)
				ORDER BY TBM_ORDER DESC
			)
			WHERE ROWNUM = 1
		)`

	for _, worker := range workers {
		if _, err := tx.ExecContext(ctx, query, worker.Jno, worker.RegUser, worker.RegNo, agent, worker.Sno, worker.UserNm, worker.Department, worker.RecordDate); err != nil {
			return utils.CustomErrorf(err)
		}
	}
	return nil
}

// 근로자 비교 반영 - 퇴직공제 등록 정보: IRIS_DEDUCTION_SET
// 선택한 프로젝트로 수정
func (r *Repository) ModifyDeductionCompareApply(ctx context.Context, tx Execer, workers entity.WorkerDailys) error {
	agent := utils.GetAgent()

	query := `
		UPDATE IRIS_DEDUCTION_SET
		SET
			JNO = :1,
			MOD_DATE = SYSDATE,
			MOD_USER = :2,
			MOD_UNO = :3,
			MOD_AGENT = :4
		WHERE ROWID = (
			SELECT ROWID FROM (
				SELECT ROWID
				FROM IRIS_DEDUCTION_SET
				WHERE SNO = :5
				AND USER_NM = :6
				AND DEPARTMENT = :7
				AND REG_NO = :8
				AND TRUNC(RECORD_DATE) = TRUNC(:9)
				ORDER BY DEDUCT_ORDER DESC
			)
			WHERE ROWNUM = 1
		)`
	for _, worker := range workers {
		if _, err := tx.ExecContext(ctx, query, worker.Jno, worker.RegUser, worker.RegUno, agent, worker.Sno, worker.UserNm, worker.Department, worker.RegNo, worker.RecordDate); err != nil {
			return utils.CustomErrorf(err)
		}
	}
	return nil
}

// 근로자 비교 반영 로그
func (r *Repository) AddCompareLog(ctx context.Context, tx Execer, logs entity.WorkerDailys) error {
	agent := utils.GetAgent()

	query := `
		INSERT INTO IRIS_COMPARE_LOG(SNO, JNO, USER_ID, USER_NM, BEFORE_STATE, AFTER_STATE, RECORD_DATE, USER_KEY, REG_DATE, REG_USER, REG_UNO, REG_AGENT)
		VALUES(:1, :2, :3, :4, :5, :6, :7, :8, SYSDATE, :9, :10, :11)`

	for _, log := range logs {
		if _, err := tx.ExecContext(ctx, query, log.Sno, log.Jno, log.UserId, log.UserNm, log.BeforeState, log.AfterState, log.RecordDate, log.UserKey, log.RegUser, log.RegUno, agent); err != nil {
			return utils.CustomErrorf(err)
		}
	}
	return nil
}
