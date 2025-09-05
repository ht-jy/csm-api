package store

import (
	"context"
	"csm-api/entity"
	"csm-api/utils"
	"database/sql"
	"errors"
	"fmt"
	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
)

/**
 * @author 작성자: 김진우
 * @created 작성일: 2025-02-17
 * @modified 최종 수정일:
 * @modifiedBy 최종 수정자:
 * @modified description
 * -
 */

// func: 전체 근로자 조회
// @param
// - page entity.PageSql: 정렬, 리스트 수
// - isRole bool : 전체프로젝트 조회 bool(true: 전체프로젝트, false: 본인이 속한 프로젝트)
// - uno string : uno를 string으로 받아 쿼리에 바로 넣음.
// - search entity.WorkerSql: 검색 단어
// - retry string: 통합검색 텍스트
func (r *Repository) GetWorkerTotalList(ctx context.Context, db Queryer, page entity.PageSql, isRole bool, uno string, search entity.Worker, retry string) (*entity.Workers, error) {
	workers := entity.Workers{}

	// 역할 조건
	roleCondition := ""
	if isRole {
		roleCondition = "AND 1 = 1"
	} else {
		roleCondition = fmt.Sprintf("AND UNO = %s", uno)
	}

	condition := ""
	condition = utils.StringWhereConvert(condition, search.JobName.NullString, "t2.JOB_NAME")
	condition = utils.StringWhereConvert(condition, search.UserId.NullString, "t1.USER_ID")
	condition = utils.StringWhereConvert(condition, search.UserNm.NullString, "t1.USER_NM")
	condition = utils.StringWhereConvert(condition, search.Department.NullString, "t1.DEPARTMENT")
	condition = utils.StringWhereConvert(condition, search.Phone.NullString, "t1.PHONE")
	condition = utils.StringWhereConvert(condition, search.WorkerType.NullString, "t1.WORKER_TYPE")
	condition = utils.StringWhereConvert(condition, search.DiscName.NullString, "t1.DISC_NAME")
	var columns []string
	columns = append(columns, "t2.JOB_NAME")
	columns = append(columns, "t1.USER_NM")
	columns = append(columns, "t1.DEPARTMENT")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	var order string
	if page.Order.Valid {
		order = page.Order.String
	} else {
		order = `
				(
					CASE 
						WHEN REG_DATE IS NULL THEN MOD_DATE 
						WHEN MOD_DATE IS NULL THEN REG_DATE 
						ELSE GREATEST(REG_DATE, MOD_DATE) 
					END
				) DESC NULLS LAST`
	}

	query := fmt.Sprintf(`
				WITH USER_IN_SNO AS (
					SELECT SNO
					FROM S_JOB_MEMBER_LIST M, IRIS_SITE_JOB J 
					WHERE 
						J.JNO = M.JNO(+)
						AND M.JNO IS NOT NULL
						%s
				),
				LATEST_DAILY AS (
					SELECT SNO, USER_KEY, MOD_DATE, REG_DATE
					FROM (
						SELECT
							SNO,
							USER_KEY,
							MOD_DATE,
							REG_DATE,
							ROW_NUMBER() OVER (
								PARTITION BY USER_KEY
								ORDER BY
									RECORD_DATE DESC,
									CASE WHEN MOD_DATE IS NOT NULL THEN 0 ELSE 1 END,
									NVL(MOD_DATE, REG_DATE) DESC
							) AS RN
						FROM IRIS_WORKER_DAILY_SET
					)
					WHERE RN = 1
				),
				JOINED AS (
					SELECT r1.*, ROW_NUMBER() OVER (PARTITION BY r1.user_nm ORDER BY r1.reg_date DESC) AS rn
					FROM IRIS_WORKER_SET R1
					LEFT JOIN LATEST_DAILY R2
					ON R1.SNO = R2.SNO AND R1.USER_KEY = R2.USER_KEY
					WHERE (R1.IS_DEL IS NULL OR R1.IS_DEL = 'N') 
				),
				BASE AS (
					SELECT *
					FROM JOINED
					WHERE rn = 1
				)
				SELECT *
				FROM (
					SELECT 
					    ROWNUM AS RNUM,
						sorted_data.SNO,
						sorted_data.SITE_NM,
						sorted_data.JNO,
						sorted_data.JOB_NAME,
						sorted_data.USER_KEY, 
						sorted_data.USER_ID, 
						sorted_data.USER_NM,
						sorted_data.DEPARTMENT,
						sorted_data.DISC_NAME,
						sorted_data.PHONE, 
						sorted_data.WORKER_TYPE,
						sorted_data.IS_RETIRE,
						sorted_data.RETIRE_DATE,
						sorted_data.IS_MANAGE,
						sorted_data.DAILY_REASON,
						sorted_data.REG_USER,
						sorted_data.REG_DATE,
						sorted_data.MOD_USER,
						sorted_data.MOD_DATE,
						sorted_data.REG_NO
					FROM (
						SELECT 
						    t1.WNO,
							t1.SNO,
							t3.SITE_NM,
							t1.JNO,
							t2.JOB_NAME,
							t1.USER_KEY,
							t1.USER_ID, 
							t1.USER_NM,
							t1.DEPARTMENT,
							t1.DISC_NAME,
							t1.PHONE, 
							t1.WORKER_TYPE,
							t1.IS_RETIRE,
							t1.RETIRE_DATE,
							t1.IS_MANAGE,
							t1.DAILY_REASON,
							t1.REG_USER,
							t1.REG_DATE,
							t1.MOD_USER,
							t1.MOD_DATE,
							COMMON.FUNC_DECODE(t1.REG_NO) AS REG_NO
						FROM BASE t1, S_JOB_INFO t2, IRIS_SITE_SET t3, USER_IN_SNO t4
						WHERE
							t1.JNO = t2.JNO(+)
							AND t1.SNO = t3.SNO(+)
							AND t3.SNO = t4.SNO
							AND t1.SNO > 100
						%s %s
						ORDER BY %s
					) sorted_data
					WHERE ROWNUM <= :1
					ORDER BY RNUM %s
				)
				WHERE RNUM > :2`, roleCondition, condition, retryCondition, order, page.RnumOrder)

	if err := db.SelectContext(ctx, &workers, query, page.EndNum, page.StartNum); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return &workers, nil
}

// func: 전체 근로자 개수 조회
// @param
// - isRole bool : 전체프로젝트 조회 bool(true: 전체프로젝트, false: 본인이 속한 프로젝트)
// - uno string : uno를 string으로 받아 쿼리에 바로 넣음.
// - searchTime string: 조회 날짜
// - retry string: 통합검색 텍스트
func (r *Repository) GetWorkerTotalCount(ctx context.Context, db Queryer, isRole bool, uno string, search entity.Worker, retry string) (int, error) {
	var count int

	roleCondition := ""
	if isRole {
		roleCondition = "AND 1 = 1"
	} else {
		roleCondition = fmt.Sprintf("AND UNO = %s", uno)
	}

	condition := ""
	condition = utils.StringWhereConvert(condition, search.JobName.NullString, "t2.JOB_NAME")
	condition = utils.StringWhereConvert(condition, search.UserId.NullString, "t1.USER_ID")
	condition = utils.StringWhereConvert(condition, search.UserNm.NullString, "t1.USER_NM")
	condition = utils.StringWhereConvert(condition, search.Department.NullString, "t1.DEPARTMENT")
	condition = utils.StringWhereConvert(condition, search.Phone.NullString, "t1.PHONE")
	condition = utils.StringWhereConvert(condition, search.WorkerType.NullString, "t1.WORKER_TYPE")

	var columns []string
	columns = append(columns, "t2.JOB_NAME")
	columns = append(columns, "t1.USER_NM")
	columns = append(columns, "t1.DEPARTMENT")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	query := fmt.Sprintf(`
						WITH USER_IN_SNO AS (
							SELECT SNO
							FROM S_JOB_MEMBER_LIST M, IRIS_SITE_JOB J 
							WHERE 
								J.JNO = M.JNO(+)
								AND M.JNO IS NOT NULL
								%s
						),
						LATEST_DAILY AS (
							SELECT SNO, USER_KEY, MOD_DATE, REG_DATE
							FROM (
								SELECT
									SNO,
									USER_KEY,
									MOD_DATE,
									REG_DATE,
									ROW_NUMBER() OVER (
										PARTITION BY USER_KEY
										ORDER BY 
											CASE WHEN MOD_DATE IS NOT NULL THEN 0 ELSE 1 END,
											NVL(MOD_DATE, REG_DATE) DESC
									) AS RN
								FROM IRIS_WORKER_DAILY_SET
							)
							WHERE RN = 1
						),
						JOINED AS (
							SELECT r1.*, ROW_NUMBER() OVER (PARTITION BY r1.user_nm ORDER BY r1.reg_date DESC) AS rn
							FROM IRIS_WORKER_SET R1
							LEFT JOIN LATEST_DAILY R2
							ON R1.SNO = R2.SNO AND R1.USER_KEY = R2.USER_KEY
							WHERE (R1.IS_DEL IS NULL OR R1.IS_DEL = 'N') 
						),
						BASE AS (
							SELECT *
							FROM JOINED
							WHERE rn = 1
						)
						SELECT 
							COUNT(*)
						FROM BASE t1, S_JOB_INFO t2, IRIS_SITE_SET t3, USER_IN_SNO t4
						WHERE
							t1.JNO = t2.JNO(+)
							AND t1.SNO = t3.SNO(+)
							AND t3.SNO = t4.SNO
							AND t1.SNO > 100
						--AND t3.IS_USE = 'Y'
						%s %s`, roleCondition, condition, retryCondition)

	if err := db.GetContext(ctx, &count, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, utils.CustomErrorf(err)
	}
	return count, nil
}

// func: 미출근 근로자 검색(현장근로자 추가시 사용)
// @param
// - userId string
func (r *Repository) GetAbsentWorkerList(ctx context.Context, db Queryer, page entity.PageSql, search entity.WorkerDaily, retry string) (*entity.Workers, error) {
	workers := entity.Workers{}

	var columns []string
	columns = append(columns, "USER_ID")
	columns = append(columns, "USER_NM")
	columns = append(columns, "DEPARTMENT")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	query := fmt.Sprintf(`
				SELECT *
				FROM (
					SELECT ROWNUM AS RNUM, sorted_data.*
					FROM (
						SELECT 
						    USER_ID, 
						    USER_NM, 
						    DEPARTMENT, 
						    :1 as RECORD_DATE,
							USER_KEY
						FROM IRIS_WORKER_SET
						WHERE JNO = :2
						AND SNO = :3
						AND USER_KEY NOT IN (
							SELECT USER_KEY
							FROM IRIS_WORKER_DAILY_SET
							WHERE JNO = :4
							AND TO_CHAR(RECORD_DATE, 'YYYY-MM-DD') = :5
						)
						%s
					) sorted_data
					WHERE ROWNUM <= :6
				)
				WHERE RNUM > :7`, retryCondition)

	if err := db.SelectContext(ctx, &workers, query, search.SearchStartTime, search.Jno, search.Sno, search.Jno, search.SearchStartTime, page.EndNum, page.StartNum); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return &workers, nil
}

// func: 미출근 근로자 개수 검색(현장근로자 추가시 사용)
// @param
// - userId string
func (r *Repository) GetAbsentWorkerCount(ctx context.Context, db Queryer, search entity.WorkerDaily, retry string) (int, error) {
	var count int

	var columns []string
	columns = append(columns, "USER_ID")
	columns = append(columns, "USER_NM")
	columns = append(columns, "DEPARTMENT")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	query := fmt.Sprintf(`
				SELECT COUNT(*)
				FROM IRIS_WORKER_SET
				WHERE JNO = :1
				AND SNO = :2
				AND USER_ID NOT IN (
					SELECT USER_KEY
					FROM IRIS_WORKER_DAILY_SET
					WHERE JNO = :3
					AND TO_CHAR(RECORD_DATE, 'YYYY-MM-DD') = :4
				)
				%s`, retryCondition)

	if err := db.GetContext(ctx, &count, query, search.Jno, search.Sno, search.Jno, search.SearchStartTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, utils.CustomErrorf(err)
	}
	return count, nil
}

// 프로젝트에 참여한 회사명 리스트
func (r *Repository) GetWorkerDepartList(ctx context.Context, db Queryer, jno int64) ([]string, error) {
	var list []string

	query := `
		SELECT DISTINCT
		  CASE
			WHEN INSTR(DEPARTMENT, ' ', -1) > 0 THEN SUBSTR(DEPARTMENT, 1, INSTR(DEPARTMENT, ' ', -1) - 1)
			ELSE DEPARTMENT
		  END AS COMPANY_NAME
		FROM IRIS_WORKER_SET
		WHERE JNO = :1
		  AND DEPARTMENT IS NOT NULL`

	if err := db.SelectContext(ctx, &list, query, jno); err != nil {
		return nil, utils.CustomErrorf(err)
	}
	return list, nil
}

// func: 근로자 추가
// @param
// -
func (r *Repository) AddWorker(ctx context.Context, tx Execer, worker entity.Worker) (int64, error) {
	agent := utils.GetAgent()

	// IRIS_WORKER_SET에 INSERT하는 쿼리
	insertQuery := `
		INSERT INTO IRIS_WORKER_SET (
			SNO, JNO, USER_ID, USER_NM, DEPARTMENT, 
			DISC_NAME, PHONE, WORKER_TYPE, IS_RETIRE, DAILY_REASON,
		    REG_DATE, REG_AGENT, REG_USER, REG_UNO, REG_NO, USER_KEY
		)
		SELECT
			:1, :2, :3, :4, :5,
			:6, REPLACE(:7, '-', ''), :8, :9, :10,
			SYSDATE, :11, :12, :13, COMMON.FUNC_ENCODE(:14), GET_IRIS_USER_UUID()
		FROM DUAL
		WHERE NOT EXISTS (
			SELECT 1
			FROM IRIS_WORKER_SET
			WHERE USER_ID = :15
			  AND USER_NM = :16
			  AND (
				(REG_NO = COMMON.FUNC_ENCODE(:17)) OR (REG_NO IS NULL AND :18 IS NULL)
			  )
			AND IS_DEL = 'N'
		)`

	res, err := tx.ExecContext(ctx, insertQuery,
		worker.Sno, worker.Jno, worker.UserId, worker.UserNm, worker.Department,
		worker.DiscName, worker.Phone, worker.WorkerType, worker.IsRetire, worker.DailyReason,
		/*, SYSDATE*/ agent, worker.RegUser, worker.RegUno, worker.RegNo,
		worker.UserId, worker.UserNm, worker.RegNo, worker.RegNo,
	)
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}
	return rows, nil
}

// func: 근로자 수정
// @param
// -
func (r *Repository) ModifyWorker(ctx context.Context, tx Execer, worker entity.Worker) error {
	agent := utils.GetAgent()

	query := `
				UPDATE IRIS_WORKER_SET R
				SET 
					R.USER_NM          = :1,
					R.DEPARTMENT       = :2,
					R.PHONE            = REPLACE(:3, '-', ''),
					R.WORKER_TYPE      = :4,
					R.IS_RETIRE        = :5,
					R.RETIRE_DATE      = :6,
					R.DAILY_REASON     = :7,
					R.MOD_DATE         = SYSDATE,
					R.MOD_AGENT        = :8,
					R.MOD_USER         = :9,
					R.MOD_UNO          = :10,
					R.TRG_EDITABLE_YN  = 'N',
					R.REG_NO           = COMMON.FUNC_ENCODE(:11),
					R.IS_MANAGE        = :12,
					R.DISC_NAME        = :13
				WHERE R.USER_KEY = :14
				  AND EXISTS (
						SELECT 1
						FROM IRIS_SITE_JOB J
						WHERE J.JNO    = R.JNO
						  AND J.IS_USE = 'Y'
				  )`

	result, err := tx.ExecContext(ctx, query,
		worker.UserNm, worker.Department, worker.Phone, worker.WorkerType, worker.IsRetire,
		worker.RetireDate, worker.DailyReason, agent, worker.ModUser, worker.ModUno, worker.RegNo,
		worker.IsManage, worker.DiscName, worker.UserKey,
	)

	if err != nil {
		return utils.CustomErrorf(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.CustomErrorf(err)
	}
	if rowsAffected == 0 {
		// 변한게 없는 경우 에러발생
		//return utils.CustomErrorf(fmt.Errorf("Rows add/update cnt: %d\n", rowsAffected))
	}

	return nil
}

// 근로자 엑셀 업로드
func (r *Repository) MergeWorker(ctx context.Context, tx Execer, worker entity.Worker) (int64, error) {
	agent := utils.GetAgent()

	query := `

		MERGE INTO IRIS_WORKER_SET W1
		USING (
			WITH worker AS (
				SELECT
					:1 AS SNO, -- 현장번호 
					:2 AS JNO, -- 프로젝트번호
					:3 AS USER_NM, -- 이름
					COMMON.FUNC_ENCODE(:4) AS REG_NO, -- 주민번호
					:5 AS USER_ID, -- 아이디
					:6 AS DEPARTMENT, -- 부서 / 조직명
					:7 AS PHONE, -- 핸드폰번호
					:8 AS DISC_NAME, -- 공종
					:9 AS IS_RETIRE, -- 퇴직여부
					(SELECT CODE FROM IRIS_CODE_SET WHERE P_CODE = 'WORKER_TYPE' AND CODE_NM = :10) AS WORKER_TYPE, -- 근로자구분
					:11 AS UNO, -- 등록자 UNO
					:12 AS NAME, -- 등록자 NAME
					:13 AS AGENT -- 등록자 AGENT
				FROM DUAL
			)
			SELECT 
				w.*,
				NVL(ws.USER_KEY, GET_IRIS_USER_UUID()) AS USER_KEY 
			FROM 
				worker w 
			LEFT JOIN 
				IRIS_WORKER_SET ws
			ON
				w.USER_ID = ws.USER_ID
				AND ws.USER_NM = w.USER_NM
				AND (
					(ws.REG_NO = w.REG_NO) OR (ws.REG_NO IS NULL AND ws.REG_NO IS NULL)
				)
				AND ws.IS_DEL = 'N'
		) W2
		ON (
			W1.USER_KEY = W2.USER_KEY
		) WHEN MATCHED THEN 
			UPDATE SET 
				W1.SNO = W2.SNO, 
				W1.JNO = W2.JNO,
				W1.REG_NO = W2.REG_NO,
				W1.DEPARTMENT = W2.DEPARTMENT,
				W1.PHONE = W2.PHONE, 
				W1.DISC_NAME = W2.DISC_NAME,
				W1.IS_RETIRE = W2.IS_RETIRE, 
				W1.WORKER_TYPE = W2.WORKER_TYPE,
				W1.MOD_DATE = SYSDATE,
				W1.MOD_UNO = W2.UNO,
				W1.MOD_USER = W2.NAME,
				W1.MOD_AGENT = W2.AGENT
		WHEN NOT MATCHED THEN
			INSERT (
				SNO, JNO, USER_ID, USER_NM, DEPARTMENT,
				DISC_NAME, PHONE, WORKER_TYPE, IS_RETIRE, IS_DEL,
				REG_DATE, REG_AGENT, REG_USER, REG_UNO, REG_NO, USER_KEY
			)
			VALUES (
				W2.SNO, W2.JNO, W2.USER_ID, W2.USER_NM, W2.DEPARTMENT,
				W2.DISC_NAME, W2.PHONE, W2.WORKER_TYPE, W2.IS_RETIRE, 'N',
				SYSDATE, W2.AGENT, W2.NAME, W2.UNO, W2.REG_NO, W2.USER_KEY
			)
		`
	res, err := tx.ExecContext(ctx, query,
		worker.Sno, worker.Jno, worker.UserNm, worker.RegNo,
		worker.UserId, worker.Department, worker.Phone, worker.DiscName,
		worker.IsRetire, worker.CodeNm, worker.RegUno, worker.RegUser, agent,
	)
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}

	return rows, nil

}

// 근로자 삭제 처리
func (r *Repository) RemoveWorker(ctx context.Context, tx Execer, worker entity.Worker) error {
	agent := utils.GetAgent()

	query := `
		UPDATE IRIS_WORKER_SET
		SET 
		    IS_DEL = 'Y',
			MOD_DATE = SYSDATE,
			MOD_AGENT = :1,
			MOD_USER = :2,
			MOD_UNO = :3
		WHERE USER_KEY = :4`

	if _, err := tx.ExecContext(ctx, query, agent, worker.ModUser, worker.ModUno, worker.UserKey); err != nil {
		return utils.CustomErrorf(err)
	}
	return nil
}

// func: 현장 근로자 조회
// @param
// - page entity.PageSql: 정렬, 리스트 수
// - search entity.WorkerSql: 검색 단어
func (r *Repository) GetWorkerSiteBaseList(ctx context.Context, db Queryer, page entity.PageSql, isRole bool, uno string, search entity.WorkerDaily, retry string) (*entity.WorkerDailys, error) {
	list := entity.WorkerDailys{}

	roleCondition := ""
	if !isRole {
		roleCondition = fmt.Sprintf("AND UNO = %s", uno)
	}

	condition := ""

	condition = utils.StringWhereConvert(condition, search.UserId.NullString, "t2.USER_ID")
	condition = utils.StringWhereConvert(condition, search.UserNm.NullString, "t2.USER_NM")
	condition = utils.StringWhereConvert(condition, search.Department.NullString, "t2.DEPARTMENT")

	var columns []string
	columns = append(columns, "t2.USER_ID")
	columns = append(columns, "t2.USER_NM")
	columns = append(columns, "t2.DEPARTMENT")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	var order string
	if page.Order.Valid {
		order = page.Order.String
	} else {
		//order = "RECORD_DATE DESC, OUT_RECOG_TIME DESC NULLS LAST"
		order = `
				RECORD_DATE DESC, (
					CASE 
						WHEN REG_DATE IS NULL THEN MOD_DATE 
						WHEN MOD_DATE IS NULL THEN REG_DATE 
						ELSE GREATEST(REG_DATE, MOD_DATE) 
					END
				) DESC NULLS LAST`
	}

	query := fmt.Sprintf(`
				WITH USER_IN_JNO AS (
					SELECT DISTINCT J.JNO
					FROM S_JOB_MEMBER_LIST M, IRIS_SITE_JOB J 
					WHERE 
						J.JNO = M.JNO(+)
						AND M.JNO IS NOT NULL
						%s
				)
				SELECT *
				FROM (
					SELECT ROWNUM AS RNUM, sorted_data.*
					FROM (
						   	SELECT 
								t1.SNO AS SNO,
								t1.JNO AS JNO,
								t1.USER_KEY AS USER_KEY,
								t2.USER_ID AS USER_ID,
								t2.USER_NM AS USER_NM,
								t2.DEPARTMENT AS DEPARTMENT,
								t1.RECORD_DATE AS RECORD_DATE,
								t1.IN_RECOG_TIME AS IN_RECOG_TIME,
								t1.OUT_RECOG_TIME AS OUT_RECOG_TIME,
								t1.IS_DEADLINE AS IS_DEADLINE,
								t1.IS_OVERTIME AS IS_OVERTIME,
								t1.REG_USER AS REG_USER,
								t1.REG_DATE AS REG_DATE,
								t1.MOD_USER AS MOD_USER,
								t1.MOD_DATE AS MOD_DATE,
								t1.WORK_STATE AS WORK_STATE,
								t1.COMPARE_STATE AS COMPARE_STATE,
								t1.WORK_HOUR as WORK_HOUR
							FROM IRIS_WORKER_DAILY_SET t1, IRIS_WORKER_SET t2, USER_IN_JNO t3
							WHERE 
							    t1.USER_KEY = t2.USER_KEY(+) 
							    AND t1.sno = t2.sno(+)
								AND t1.jno = t3.jno
								AND t1.SNO > 100
								AND T2.IS_DEL = 'N'
								AND t1.COMPARE_STATE in ('S', 'X')
								AND t1.JNO = :1
								AND TO_CHAR(t1.RECORD_DATE, 'yyyy-mm-dd') BETWEEN :2 AND :3
							%s %s
							ORDER BY %s
					) sorted_data
					WHERE ROWNUM <= :4
					ORDER BY RNUM %s
				)
				WHERE RNUM > :5`, roleCondition, condition, retryCondition, order, page.RnumOrder)

	if err := db.SelectContext(ctx, &list, query, search.Jno, search.SearchStartTime, search.SearchEndTime, page.EndNum, page.StartNum); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return &list, nil
}

// func: 현장 근로자 개수 조회
// @param
// - searchTime string: 조회 날짜
func (r *Repository) GetWorkerSiteBaseCount(ctx context.Context, db Queryer, isRole bool, uno string, search entity.WorkerDaily, retry string) (int, error) {
	var count int

	roleCondition := ""
	if !isRole {
		roleCondition = fmt.Sprintf("AND UNO = %s", uno)
	}

	condition := ""

	condition = utils.StringWhereConvert(condition, search.UserId.NullString, "t2.USER_ID")
	condition = utils.StringWhereConvert(condition, search.UserNm.NullString, "t2.USER_NM")
	condition = utils.StringWhereConvert(condition, search.Department.NullString, "t2.DEPARTMENT")

	var columns []string
	columns = append(columns, "t2.USER_ID")
	columns = append(columns, "t2.USER_NM")
	columns = append(columns, "t2.DEPARTMENT")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	query := fmt.Sprintf(`
							WITH USER_IN_JNO AS (
								SELECT DISTINCT J.JNO
								FROM S_JOB_MEMBER_LIST M, IRIS_SITE_JOB J 
								WHERE 
									J.JNO = M.JNO(+)
									AND M.JNO IS NOT NULL
									%s
							)
							SELECT 
								count(*)
							FROM IRIS_WORKER_DAILY_SET t1, IRIS_WORKER_SET t2, USER_IN_JNO t3
							WHERE 
							    t1.USER_KEY = t2.USER_KEY(+) 
							    AND t1.sno = t2.sno(+)
								AND t1.jno = t3.jno
								AND t1.SNO > 100
								AND T2.IS_DEL = 'N'
								AND t1.COMPARE_STATE in ('S', 'X')
								AND t1.JNO = :1
								AND TO_CHAR(t1.RECORD_DATE, 'yyyy-mm-dd') BETWEEN :2 AND :3
							%s %s`, roleCondition, condition, retryCondition)

	if err := db.GetContext(ctx, &count, query, search.Jno, search.SearchStartTime, search.SearchEndTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, utils.CustomErrorf(err)
	}
	return count, nil
}

// func: 현장 근로자 조회 - 협력업체
// @param
// - page entity.PageSql: 정렬, 리스트 수
// - search entity.WorkerSql: 검색 단어
func (r *Repository) GetWorkerSiteBaseListByCompany(ctx context.Context, db Queryer, page entity.PageSql, id string, search entity.WorkerDaily, retry string) (*entity.WorkerDailys, error) {
	list := entity.WorkerDailys{}

	roleCondition := fmt.Sprintf("AND ID = %s", id)

	condition := ""

	condition = utils.StringWhereConvert(condition, search.UserId.NullString, "t2.USER_ID")
	condition = utils.StringWhereConvert(condition, search.UserNm.NullString, "t2.USER_NM")
	condition = utils.StringWhereConvert(condition, search.Department.NullString, "t2.DEPARTMENT")

	var columns []string
	columns = append(columns, "t2.USER_ID")
	columns = append(columns, "t2.USER_NM")
	columns = append(columns, "t2.DEPARTMENT")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	var order string
	if page.Order.Valid {
		order = page.Order.String
	} else {
		//order = "RECORD_DATE DESC, OUT_RECOG_TIME DESC NULLS LAST"
		order = `
				RECORD_DATE DESC, (
					CASE 
						WHEN REG_DATE IS NULL THEN MOD_DATE 
						WHEN MOD_DATE IS NULL THEN REG_DATE 
						ELSE GREATEST(REG_DATE, MOD_DATE) 
					END
				) DESC NULLS LAST`
	}

	query := fmt.Sprintf(`
				WITH USER_IN_JNO AS (
					SELECT DISTINCT J.JNO, S.COMP_NAME
					FROM JOB_SUBCON_INFO S, IRIS_SITE_JOB J 
					WHERE 
						J.JNO = S.JNO(+)
						AND S.JNO IS NOT NULL
						%s
				)
				SELECT *
				FROM (
					SELECT ROWNUM AS RNUM, sorted_data.*
					FROM (
						   	SELECT 
								t1.SNO AS SNO,
								t1.JNO AS JNO,
								t1.USER_KEY AS USER_KEY,
								t2.USER_ID AS USER_ID,
								t2.USER_NM AS USER_NM,
								t2.DEPARTMENT AS DEPARTMENT,
								t1.RECORD_DATE AS RECORD_DATE,
								t1.IN_RECOG_TIME AS IN_RECOG_TIME,
								t1.OUT_RECOG_TIME AS OUT_RECOG_TIME,
								t1.IS_DEADLINE AS IS_DEADLINE,
								t1.IS_OVERTIME AS IS_OVERTIME,
								t1.REG_USER AS REG_USER,
								t1.REG_DATE AS REG_DATE,
								t1.MOD_USER AS MOD_USER,
								t1.MOD_DATE AS MOD_DATE,
								t1.WORK_STATE AS WORK_STATE,
								t1.COMPARE_STATE AS COMPARE_STATE,
								t1.WORK_HOUR as WORK_HOUR
							FROM IRIS_WORKER_DAILY_SET t1, IRIS_WORKER_SET t2, USER_IN_JNO t3
							WHERE 
							    t1.USER_KEY = t2.USER_KEY(+) 
							    AND t1.sno = t2.sno(+)
								AND t1.jno = t3.jno
								AND t1.SNO > 100
								AND t2.DEPARTMENT LIKE '%%' || TRIM(REPLACE(NVL(t3.COMP_NAME, ''), '주식회사', '')) || '%%'
								AND T2.IS_DEL = 'N'
								AND t1.COMPARE_STATE in ('S', 'X')
								AND t1.JNO = :1
								AND TO_CHAR(t1.RECORD_DATE, 'yyyy-mm-dd') BETWEEN :2 AND :3
							%s %s
							ORDER BY %s
					) sorted_data
					WHERE ROWNUM <= :4
					ORDER BY RNUM %s
				)
				WHERE RNUM > :5`, roleCondition, condition, retryCondition, order, page.RnumOrder)

	if err := db.SelectContext(ctx, &list, query, search.Jno, search.SearchStartTime, search.SearchEndTime, page.EndNum, page.StartNum); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return &list, nil
}

// func: 현장 근로자 개수 조회 - 협력업체
// @param
// - searchTime string: 조회 날짜
func (r *Repository) GetWorkerSiteBaseByCompanyCount(ctx context.Context, db Queryer, id string, search entity.WorkerDaily, retry string) (int, error) {
	var count int

	roleCondition := fmt.Sprintf("AND ID = %s", id)

	condition := ""

	condition = utils.StringWhereConvert(condition, search.UserId.NullString, "t2.USER_ID")
	condition = utils.StringWhereConvert(condition, search.UserNm.NullString, "t2.USER_NM")
	condition = utils.StringWhereConvert(condition, search.Department.NullString, "t2.DEPARTMENT")

	var columns []string
	columns = append(columns, "t2.USER_ID")
	columns = append(columns, "t2.USER_NM")
	columns = append(columns, "t2.DEPARTMENT")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	query := fmt.Sprintf(`
							WITH USER_IN_JNO AS (
								SELECT DISTINCT J.JNO, S.COMP_NAME
								FROM JOB_SUBCON_INFO S, IRIS_SITE_JOB J 
								WHERE 
									J.JNO = S.JNO(+)
									AND S.JNO IS NOT NULL
									%s
							)
							SELECT 
								count(*)
							FROM IRIS_WORKER_DAILY_SET t1, IRIS_WORKER_SET t2, USER_IN_JNO t3
							WHERE 
							    t1.USER_KEY = t2.USER_KEY(+) 
							    AND t1.sno = t2.sno(+)
								AND t1.jno = t3.jno
								AND t1.SNO > 100
								AND t2.DEPARTMENT LIKE '%%' || TRIM(REPLACE(NVL(t3.COMP_NAME, ''), '주식회사', '')) || '%%'
								AND T2.IS_DEL = 'N'
								AND t1.COMPARE_STATE in ('S', 'X')
								AND t1.JNO = :1
								AND TO_CHAR(t1.RECORD_DATE, 'yyyy-mm-dd') BETWEEN :2 AND :3
							%s %s`, roleCondition, condition, retryCondition)

	if err := db.GetContext(ctx, &count, query, search.Jno, search.SearchStartTime, search.SearchEndTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, utils.CustomErrorf(err)
	}
	return count, nil
}

// func: 현장 근로자 추가/수정
// @param
// -
func (r *Repository) MergeSiteBaseWorker(ctx context.Context, tx Execer, workers entity.WorkerDailys) error {
	agent := utils.GetAgent()

	query := `
				MERGE INTO IRIS_WORKER_DAILY_SET t1
				USING (
					SELECT 
						:1 AS SNO,
						:2 AS JNO,
						:3 AS USER_KEY,
						:4 AS RECORD_DATE,
						:5 AS IN_RECOG_TIME,
						:6 AS OUT_RECOG_TIME,
						:7 AS REG_AGENT,
						:8 AS REG_USER,
						:9 AS REG_UNO,
						:10 AS IS_DEADLINE,
						:11 AS WORK_STATE,
						:12 AS IS_OVERTIME,
						:13 AS WORK_HOUR
					FROM DUAL
				) t2
				ON (
					t1.SNO = t2.SNO 
					AND t1.JNO = t2.JNO 
					AND t1.USER_KEY = t2.USER_KEY
				    AND t1.RECORD_DATE   = t2.RECORD_DATE
				) WHEN MATCHED THEN
					UPDATE SET
						t1.IN_RECOG_TIME = t2.IN_RECOG_TIME,
						t1.OUT_RECOG_TIME = t2.OUT_RECOG_TIME,
						t1.MOD_DATE      = SYSDATE,
						t1.MOD_AGENT     = t2.REG_AGENT,
						t1.MOD_USER      = t2.REG_USER,
						t1.MOD_UNO       = t2.REG_UNO,
				    	t1.IS_DEADLINE   = t2.IS_DEADLINE,
				    	t1.WORK_STATE = t2.WORK_STATE,
						t1.IS_OVERTIME   = t2.IS_OVERTIME,
						t1.WORK_HOUR   = t2.WORK_HOUR
					WHERE t1.SNO = t2.SNO
					AND t1.JNO = t2.JNO
					AND t1.USER_KEY = t2.USER_KEY
				    AND t1.RECORD_DATE   = t2.RECORD_DATE
				WHEN NOT MATCHED THEN
					INSERT (SNO, JNO, USER_KEY, RECORD_DATE, IN_RECOG_TIME, OUT_RECOG_TIME, WORK_STATE, COMPARE_STATE, WORK_HOUR, REG_DATE, REG_AGENT, REG_USER, REG_UNO, IS_DEADLINE, IS_OVERTIME)
					VALUES (t2.SNO, t2.JNO, t2.USER_KEY, t2.RECORD_DATE, t2.IN_RECOG_TIME, t2.OUT_RECOG_TIME, t2.WORK_STATE, 'X', t2.WORK_HOUR, SYSDATE, t2.REG_AGENT, t2.REG_USER, t2.REG_UNO, t2.IS_DEADLINE, t2.IS_OVERTIME)`

	for _, worker := range workers {
		_, err := tx.ExecContext(ctx, query,
			worker.Sno, worker.Jno, worker.UserKey, worker.RecordDate, worker.InRecogTime,
			worker.OutRecogTime, agent, worker.ModUser, worker.ModUno, worker.IsDeadline,
			worker.WorkState, worker.IsOvertime, worker.WorkHour,
		)
		if err != nil {
			return utils.CustomErrorf(err)
		}
	}

	return nil
}

// 현장 근로자 변경사항 로그 저장
func (r *Repository) MergeSiteBaseWorkerLog(ctx context.Context, tx Execer, workers entity.WorkerDailys) error {
	agent := utils.GetAgent()

	query := `
		INSERT INTO IRIS_WORKER_DAILY_LOG(SNO, JNO, USER_ID, RECOG_TIME, TRANS_TYPE, MESSAGE, USER_KEY, REG_DATE, REG_USER, REG_UNO, REG_AGENT)
		VALUES(:1, :2, :3, :4, :5, :6, :7, SYSDATE, :8, :9, :10)`

	for _, worker := range workers {
		if _, err := tx.ExecContext(ctx, query, worker.Sno, worker.Jno, worker.UserId, worker.RecordDate, worker.WorkState, worker.Message, worker.UserKey, worker.ModUser, worker.ModUno, agent); err != nil {
			return utils.CustomErrorf(err)
		}
	}
	return nil
}

// func: 현장 근로자 일괄마감
// @param
// -
func (r *Repository) ModifyWorkerDeadline(ctx context.Context, tx Execer, workers entity.WorkerDailys) error {
	agent := utils.GetAgent()

	query := `
				UPDATE IRIS_WORKER_DAILY_SET 
				SET 
					IS_DEADLINE = 'Y',
					MOD_DATE = SYSDATE,
					MOD_AGENT = :1,
					MOD_USER = :2,
					MOD_UNO = :3
				WHERE SNO = :4
				AND JNO = :5
				AND USER_KEY = :6
				AND RECORD_DATE = :7`

	for _, worker := range workers {
		_, err := tx.ExecContext(ctx, query,
			agent, worker.ModUser, worker.ModUno, worker.Sno, worker.Jno,
			worker.UserKey, worker.RecordDate,
		)
		if err != nil {
			return utils.CustomErrorf(err)
		}
	}

	return nil
}

// func: 현장 근로자 프로젝트 변경
// @param
// -
func (r *Repository) ModifyWorkerProject(ctx context.Context, tx Execer, workers entity.WorkerDailys) error {
	agent := utils.GetAgent()

	query := `
				UPDATE IRIS_WORKER_DAILY_SET 
				SET 
				    JNO = :1,
					MOD_DATE = SYSDATE,
					MOD_AGENT = :2,
					MOD_USER = :3,
					MOD_UNO = :4
				WHERE SNO = :5
				AND JNO = :6
				AND USER_KEY = :7
				AND RECORD_DATE = :8`

	for _, worker := range workers {
		_, err := tx.ExecContext(ctx, query,
			worker.AfterJno, agent, worker.ModUser, worker.ModUno, worker.Sno,
			worker.Jno, worker.UserKey, worker.RecordDate,
		)
		if err != nil {
			return utils.CustomErrorf(err)
		}
	}

	return nil
}

// 현장 근로자 프로젝트 변경시 같은 현장내 프로젝트일 경우 전체 근로자 프로젝트 변경
func (r *Repository) ModifyWorkerDefaultProject(ctx context.Context, tx Execer, workers entity.WorkerDailys) error {
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
			AND EXISTS (
				SELECT 1
				FROM IRIS_SITE_JOB
				WHERE SNO = :7 AND JNO = :8
			)`

	for _, worker := range workers {
		if _, err := tx.ExecContext(ctx, query, worker.AfterJno, worker.ModUser, worker.ModUno, agent, worker.Sno, worker.UserKey, worker.Sno, worker.Jno); err != nil {
			return utils.CustomErrorf(err)
		}
	}
	return nil
}

// func: 현장 근로자 일일 마감처리
// @param
// -
func (r *Repository) ModifyWorkerDeadlineInit(ctx context.Context, tx Execer) error {
	agent := utils.GetAgent()

	query := `
			UPDATE IRIS_WORKER_DAILY_SET 
			SET 
				IS_DEADLINE = 'Y',
				MOD_DATE = SYSDATE,
				MOD_AGENT = :1,
				MOD_USER = 'Scheduled'
			WHERE TRUNC(RECORD_DATE) >= TRUNC(SYSDATE) - 7
			AND TRUNC(RECORD_DATE) < TRUNC(SYSDATE)
			AND WORK_STATE = '02'
			AND IS_DEADLINE = 'N'
			AND COMPARE_STATE = 'S'`

	if _, err := tx.ExecContext(ctx, query, agent); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}

// func: 철야 근로자 조회
// @param
// -
func (r *Repository) GetWorkerOverTime(ctx context.Context, db Queryer) (*entity.WorkerOverTimes, error) {

	workerOverTimes := entity.WorkerOverTimes{}
	query := `
			SELECT 
				w1.CNO AS BEFORE_CNO, 
				w2.OUT_RECOG_TIME AS OUT_RECOG_TIME, 
				w2.CNO AS AFTER_CNO 
			FROM iris_worker_daily_set w1 
			INNER JOIN iris_worker_daily_set w2 
			ON w1.user_id = w2.user_id AND w1.jno = w2.jno 
			WHERE to_date(w2.record_date) = TRUNC(SYSDATE) 
			  AND w2.IN_RECOG_TIME IS NULL 
			  AND w2.OUT_RECOG_TIME IS NOT NULL 
			  AND TO_DATE(w1.RECORD_DATE) = TRUNC(SYSDATE - 1) 
			  AND w1.IN_RECOG_TIME IS NOT NULL 
			  AND w1.OUT_RECOG_TIME IS NULL
			  AND W2.COMPARE_STATE = 'S'
		`

	if err := db.SelectContext(ctx, &workerOverTimes, query); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return &workerOverTimes, nil

}

// func: 현장 근로자 철야 처리
// @param
// - workerOverTime entity.WorkerOverTime: BeforeCno, AfterCno, OutRecogTime
func (r *Repository) ModifyWorkerOverTime(ctx context.Context, tx Execer, workerOverTime entity.WorkerOverTime) error {
	agent := utils.GetAgent()

	query := `
		UPDATE 
		    IRIS_WORKER_DAILY_SET 
		SET 
		    OUT_RECOG_TIME = :1,
		    IS_OVERTIME = 'Y',
		    WORK_STATE = '02',
			MOD_DATE = SYSDATE,
			MOD_AGENT = :2,
			MOD_USER = 'Scheduled'
		WHERE 
		    CNO = :3
			
	`

	if _, err := tx.ExecContext(ctx, query, workerOverTime.OutRecogTime, agent, workerOverTime.BeforeCno); err != nil {
		return utils.CustomErrorf(err)
	}
	return nil

}

// func: 현장 근로자 철야 처리 후 삭제
// @param
// - cno: 근로자 PK
func (r *Repository) DeleteWorkerOverTime(ctx context.Context, tx Execer, cno null.Int) error {

	query := `
		DELETE FROM iris_worker_daily_set
		WHERE  CNO = :1
		`
	if _, err := tx.ExecContext(ctx, query, cno); err != nil {
		return utils.CustomErrorf(err)
	}
	return nil
}

// 현장 근로자 삭제
func (r *Repository) RemoveSiteBaseWorkers(ctx context.Context, tx Execer, workers entity.WorkerDailys) error {
	query := `
		DELETE FROM IRIS_WORKER_DAILY_SET
		WHERE SNO = :1
		AND JNO = :2
		AND USER_KEY = :3
		AND TRUNC(RECORD_DATE) = TRUNC(:4)
		AND IS_DEADLINE = 'N'`

	for _, worker := range workers {
		if _, err := tx.ExecContext(ctx, query, worker.Sno, worker.Jno, worker.UserKey, worker.RecordDate); err != nil {
			return utils.CustomErrorf(err)
		}
	}
	return nil
}

// 마감 취소
func (r *Repository) ModifyDeadlineCancel(ctx context.Context, tx Execer, workers entity.WorkerDailys) error {
	query := `
		UPDATE IRIS_WORKER_DAILY_SET
		SET
			IS_DEADLINE = 'N',
			MOD_DATE = SYSDATE,
			MOD_USER = :1,
			MOD_UNO = :2
		WHERE SNO = :3
		AND JNO = :4
		AND USER_KEY = :5
		AND TRUNC(RECORD_DATE) = TRUNC(:6)`

	for _, worker := range workers {
		if _, err := tx.ExecContext(ctx, query, worker.ModUser, worker.ModUno, worker.Sno, worker.Jno, worker.UserKey, worker.RecordDate); err != nil {
			return utils.CustomErrorf(err)
		}
	}
	return nil
}

// 현장 근로자 근로자 키 조회
func (r *Repository) GetDailyWorkerUserKey(ctx context.Context, db Queryer, worker entity.WorkerDaily) (string, error) {
	var userKey string
	query := `
		SELECT USER_KEY
		FROM IRIS_WORKER_SET
		WHERE REPLACE(PHONE, '-', '') = :1
		AND USER_NM = :2
		AND SUBSTR(COMMON.FUNC_DECODE(REG_NO), 1, 6) = :3
		AND SNO = :4
		AND JNO = :5`
	if err := db.GetContext(ctx, &userKey, query, worker.Phone, worker.UserNm, worker.RegNo, worker.Sno, worker.Jno); err != nil {
		return "", utils.CustomErrorf(err)
	}
	return userKey, nil
}

// 현장근로자 추가
func (r *Repository) AddDailyWorkers(ctx context.Context, db Queryer, tx Execer, workers entity.WorkerDailys) (entity.WorkerDailys, error) {
	agent := utils.GetAgent()

	insertQuery := `
		MERGE INTO IRIS_WORKER_DAILY_SET T
		USING (
			SELECT
				:1  AS SNO,
				:2  AS JNO,
				:3  AS USER_KEY,
				:4  AS RECORD_DATE,
				:5  AS IN_RECOG_TIME,
				:6  AS OUT_RECOG_TIME,
				:7  AS WORK_STATE,
				:8  AS COMPARE_STATE,
				:9  AS WORK_HOUR,
				:10 AS REG_USER,
				:11 AS REG_UNO,
				:12 AS REG_AGENT
			FROM DUAL
		) SRC
		ON (
			T.SNO = SRC.SNO
			AND T.USER_KEY = SRC.USER_KEY
			AND T.RECORD_DATE = SRC.RECORD_DATE
		)
		WHEN MATCHED THEN
			UPDATE SET
				T.IN_RECOG_TIME  = SRC.IN_RECOG_TIME,
				T.OUT_RECOG_TIME = SRC.OUT_RECOG_TIME,
				T.WORK_STATE     = SRC.WORK_STATE,
				T.COMPARE_STATE  = SRC.COMPARE_STATE,
				T.WORK_HOUR      = SRC.WORK_HOUR,
				T.MOD_DATE       = SYSDATE,
				T.MOD_USER       = SRC.REG_USER,
				T.MOD_UNO        = SRC.REG_UNO,
				T.MOD_AGENT      = SRC.REG_AGENT
		WHEN NOT MATCHED THEN
			INSERT (
				SNO, JNO, USER_KEY, RECORD_DATE, IN_RECOG_TIME,
				OUT_RECOG_TIME, WORK_STATE, COMPARE_STATE, WORK_HOUR, REG_DATE,
				REG_USER, REG_UNO, REG_AGENT
			) VALUES (
				SRC.SNO, SRC.JNO, SRC.USER_KEY, SRC.RECORD_DATE, SRC.IN_RECOG_TIME,
				SRC.OUT_RECOG_TIME, SRC.WORK_STATE, SRC.COMPARE_STATE, SRC.WORK_HOUR, SYSDATE,
				SRC.REG_USER, SRC.REG_UNO, SRC.REG_AGENT
			)
	`

	var insertedWorkers entity.WorkerDailys

	for _, worker := range workers {
		_, err := tx.ExecContext(ctx, insertQuery,
			worker.Sno, worker.Jno, worker.UserKey, worker.RecordDate, worker.InRecogTime,
			worker.OutRecogTime, worker.WorkState, worker.CompareState, worker.WorkHour, worker.RegUser,
			worker.RegUno, agent,
		)
		if err != nil {
			return nil, utils.CustomErrorf(err)
		}
		// 통과된 항목만 슬라이스에 append
		copied := worker // 새 인스턴스를 만들어야 주소 복사 문제 방지됨
		copied.ModUser = worker.RegUser
		copied.ModUno = worker.RegUno
		copied.Message = utils.ParseNullString(fmt.Sprintf("[ADD DATA]in_recog_time: %v|out_recog_time: %v|work_hour: %v",
			worker.InRecogTime.Time.Format("15:04:05"),
			worker.OutRecogTime.Time.Format("15:04:05"),
			worker.WorkHour.Float64,
		))
		insertedWorkers = append(insertedWorkers, copied)
	}

	return insertedWorkers, nil
}

// 프로젝트, 기간내 모든 현장근로자 근태정보 조회
func (r *Repository) GetDailyWorkersByJnoAndDate(ctx context.Context, db Queryer, param entity.RecordDailyWorkerReq) ([]entity.RecordDailyWorkerRes, error) {
	var list []entity.RecordDailyWorkerRes

	query := `
		SELECT 
			T3.JOB_NAME,
			T1.USER_NM,
			T1.DEPARTMENT,
			T1.USER_ID AS PHONE,
			T2.RECORD_DATE,
			T2.IN_RECOG_TIME,
			T2.OUT_RECOG_TIME,
			T2.WORK_HOUR,
			T2.IS_DEADLINE 
		FROM IRIS_WORKER_SET T1
		LEFT JOIN IRIS_WORKER_DAILY_SET T2 ON T1.SNO = T2.SNO AND T1.USER_KEY = T2.USER_KEY
		LEFT JOIN S_JOB_INFO T3 ON T2.JNO = T3.JNO
		WHERE T2.JNO = :1
		AND TO_CHAR(T2.RECORD_DATE, 'yyyy-mm-dd') BETWEEN :2 AND :3
		AND T2.COMPARE_STATE IN ('S', 'X')`

	if err := db.SelectContext(ctx, &list, query, param.Jno, param.StartDate, param.EndDate); err != nil {
		return list, utils.CustomErrorf(err)
	}
	return list, nil
}

// 현장근로자 일괄 공수 변경
func (r *Repository) ModifyWorkHours(ctx context.Context, tx Execer, workers entity.WorkerDailys) error {
	agent := utils.GetAgent()

	query := `
		UPDATE IRIS_WORKER_DAILY_SET
		SET
			WORK_HOUR = :1,
			MOD_DATE = SYSDATE,
			MOD_AGENT = :2,
			MOD_USER = :3,
			MOD_UNO = :4
		WHERE SNO = :5
		AND JNO = :6
		AND USER_KEY = :7
		AND RECORD_DATE = :8`

	for _, worker := range workers {
		if _, err := tx.ExecContext(ctx, query, worker.WorkHour, agent, worker.ModUser, worker.ModUno, worker.Sno, worker.Jno, worker.UserKey, worker.RecordDate); err != nil {
			return utils.CustomErrorf(err)
		}
	}
	return nil
}

// 홍채인식기 IRIS_WORKER_SET 테이블 미기록 리스트 조회::스케줄 용도
func (r *Repository) GetRecdWorkerList(ctx context.Context, db Queryer) ([]entity.Worker, error) {
	var list []entity.Worker

	query := `
		SELECT 
			IRIS_NO, 
			SNO, 
			JNO, 
			USER_ID, 
			USER_NM, 
			NVL(substr(TRIM(DEPARTMENT), 0, INSTR(TRIM(DEPARTMENT), ' ', -1)), DEPARTMENT) AS DEPARTMENT, 
			DISC_NAME, 
			COMMON.FUNC_DECODE(REG_NO) AS REG_NO,
			CASE 
				WHEN INSTR(NVL(DEPARTMENT, ''), '하이테크') > 0 OR INSTR(NVL(UPPER(DEPARTMENT), ''), 'HTENC') > 0 THEN '01'
			    WHEN INSTR(NVL(DEPARTMENT, ''), '관리') > 0 OR INSTR(NVL(DISC_NAME, ''), '관리') > 0 THEN '02'
				ELSE '00'
			END AS WORKER_TYPE,
			CASE 
				WHEN INSTR(NVL(DEPARTMENT, ''), '관리') > 0 OR INSTR(NVL(DISC_NAME, ''), '관리') > 0 THEN 'Y'
				ELSE 'N'
			END AS IS_MANAGE,
		    CASE 
				WHEN INSTR(NVL(DEPARTMENT, ''), '퇴사') > 0 OR INSTR(NVL(DISC_NAME, ''), '퇴사') > 0 THEN 'Y'
				ELSE 'N'
			END AS IS_RETIRE
		FROM IRIS_RECD_SET
		WHERE IS_WORKER = 'N'`

	if err := db.SelectContext(ctx, &list, query); err != nil {
		return list, utils.CustomErrorf(err)
	}
	return list, nil
}

// user_key 조회::스케줄 용도
func (r *Repository) GetRecdWorkerUserKey(ctx context.Context, db Queryer, worker entity.Worker) (string, error) {
	var userKey string

	query := `
		SELECT USER_KEY
		FROM (
			SELECT USER_KEY 
			FROM IRIS_WORKER_SET
			WHERE USER_ID = :1 
			AND USER_NM = :2
			AND (
				COMMON.FUNC_DECODE(REG_NO) = :3
				OR (REG_NO IS NULL AND :4 IS NULL)
			) AND IS_DEL = 'N'
			ORDER BY REG_DATE DESC
		)
		WHERE ROWNUM = 1`

	if err := db.GetContext(ctx, &userKey, query, worker.UserId, worker.UserNm, worker.RegNo, worker.RegNo); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			query = `SELECT GET_IRIS_USER_UUID() AS USER_KEY FROM DUAL`
			if err = db.GetContext(ctx, &userKey, query); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return "", utils.CustomErrorf(fmt.Errorf("user_key not found"))
				}
				return userKey, utils.CustomErrorf(err)
			}
			return userKey, nil
		}
		return userKey, utils.CustomErrorf(err)
	}
	return userKey, nil
}

// 홍채인식기 데이터를 근로자 테이블에 반영::스케줄 용도
func (r *Repository) MergeRecdWorker(ctx context.Context, tx Execer, worker []entity.Worker) error {
	agent := utils.GetAgent()

	query := `
		MERGE INTO IRIS_WORKER_SET t1
		USING (
			SELECT 
				:1 AS USER_KEY, 
				:2 AS SNO, 
				:3 AS JNO, 
				:4 AS USER_ID, 
				:5 AS USER_NM,
				:6 AS DEPARTMENT, 
				:7 AS DISC_NAME, 
				COMMON.FUNC_ENCODE(:8) AS REG_NO,
				:9 AS WORKER_TYPE, 
				:10 AS IS_MANAGE,
				:11 AS IS_RETIRE,
				:12 AS MOD_AGENT,
				0 AS MOD_UNO
			FROM DUAL
		) t2
		ON (
			t1.USER_KEY = t2.USER_KEY AND t1.SNO = t2.SNO
		)
		WHEN MATCHED THEN
			UPDATE SET 
				t1.JNO = t2.JNO, 
				t1.USER_ID = t2.USER_ID, 
				t1.USER_NM = t2.USER_NM,
				t1.DEPARTMENT = t2.DEPARTMENT, 
				t1.WORKER_TYPE = t2.WORKER_TYPE,
				t1.IS_MANAGE = t2.IS_MANAGE,
				t1.IS_RETIRE = t2.IS_RETIRE,
				t1.PHONE = t2.USER_ID, 
				t1.DISC_NAME = t2.DISC_NAME,
				t1.REG_NO = t2.REG_NO,
				t1.MOD_DATE = SYSDATE, 
				t1.MOD_USER = 'TRG_IRIS_WORKER_SET',
				t1.MOD_UNO = t2.MOD_UNO,
				t1.MOD_AGENT = t2.MOD_AGENT
			WHERE (t1.TRG_EDITABLE_YN = 'Y' OR t1.TRG_EDITABLE_YN IS NULL)
		WHEN NOT MATCHED THEN
			INSERT (
				USER_KEY, SNO, JNO, USER_ID, USER_NM,
				DEPARTMENT, WORKER_TYPE, IS_MANAGE, IS_RETIRE, DISC_NAME, REG_NO, 
				REG_DATE, REG_USER, REG_UNO, REG_AGENT
			) VALUES (
				t2.USER_KEY, t2.SNO, t2.JNO, t2.USER_ID, t2.USER_NM, 
				t2.DEPARTMENT, t2.WORKER_TYPE, t2.IS_MANAGE, t2.IS_RETIRE, t2.DISC_NAME, t2.REG_NO, 
				SYSDATE, 'TRG_IRIS_WORKER_SET', t2.MOD_UNO, t2.MOD_AGENT
			)`

	query2 := `
		UPDATE IRIS_RECD_SET
			SET IS_WORKER = 'Y'
		WHERE IRIS_NO = :1`

	for _, w := range worker {
		if _, err := tx.ExecContext(ctx, query, w.UserKey, w.Sno, w.Jno, w.UserId, w.UserNm, w.Department, w.DiscName, w.RegNo, w.WorkerType, w.IsManage, w.IsRetire, agent); err != nil {
			return utils.CustomErrorf(err)
		}

		if _, err := tx.ExecContext(ctx, query2, w.IrisNo); err != nil {
			return utils.CustomErrorf(err)
		}
	}

	return nil
}

// 홍채인식기 데이터 현장근로자(IRIS_WORKER_DAILY_SET) 미반영 조회
func (r *Repository) GetRecdDailyWorkerList(ctx context.Context, db Queryer) ([]entity.WorkerDaily, error) {
	var list []entity.WorkerDaily

	query := `
		SELECT IRIS_NO, DNO, SNO, JNO, USER_ID, USER_NM, COMMON.FUNC_DECODE(REG_NO) AS REG_NO, RECOG_TIME AS RECORD_DATE
		FROM IRIS_RECD_SET
		WHERE IS_WORKER = 'Y'
		AND IS_DAILY_WORKER = 'N'`

	if err := db.SelectContext(ctx, &list, query); err != nil {
		return list, utils.CustomErrorf(err)
	}
	return list, nil
}

// 출근 기록이 있는지 확인
func (r *Repository) GetRecdDailyWorkerChk(ctx context.Context, db Queryer, userKey string, date null.Time) (bool, error) {
	var chk bool

	qeury := `
		SELECT 1
		FROM IRIS_WORKER_DAILY_SET
		WHERE USER_KEY = :1
		AND TRUNC(RECORD_DATE) = :2`
	if err := db.GetContext(ctx, &chk, qeury, userKey, date); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, utils.CustomErrorf(err)
	}
	return true, nil
}

// 홍채인식기 데이터 현장근로자(IRIS_WORKER_DAILY_SET) 테이블에 반영
func (r *Repository) MergeRecdDailyWorker(ctx context.Context, tx Execer, worker []entity.WorkerDaily) error {
	agent := utils.GetAgent()

	query := `
		MERGE INTO IRIS_WORKER_DAILY_SET t1
		USING (
			SELECT 
				:1 AS SNO, 
				:2 AS JNO, 
				:3 AS USER_KEY,
				TRUNC(:4) AS RECORD_DATE,
				:5 AS IN_RECOG_TIME,
				:6 AS OUT_RECOG_TIME,
				:7 AS WORK_STATE,
				0 AS MOD_UNO,
				:8 AS MOD_AGENT,
				:9 AS DNO	
			FROM DUAL
		) t2
		ON (
			t1.USER_KEY = t2.USER_KEY
			AND t1.SNO = t2.SNO
			AND t1.RECORD_DATE = t2.RECORD_DATE
		)
		WHEN MATCHED THEN
			UPDATE SET
				t1.OUT_RECOG_TIME = t2.OUT_RECOG_TIME,
				t1.WORK_STATE = t2.WORK_STATE,
				t1.MOD_DATE = SYSDATE,
				t1.MOD_USER = 'TRG_IRIS_WORKER_DAILY_SET',
				t1.MOD_UNO = t2.MOD_UNO,
				t1.MOD_AGENT = t2.MOD_AGENT,
				t1.DNO = t2.DNO
			WHERE t2.OUT_RECOG_TIME IS NOT NULL
			AND (t1.OUT_RECOG_TIME IS NULL OR t2.OUT_RECOG_TIME > t1.OUT_RECOG_TIME)
		WHEN NOT MATCHED THEN
			INSERT (
				SNO, JNO, USER_KEY, RECORD_DATE, IN_RECOG_TIME, 
				OUT_RECOG_TIME, WORK_STATE, REG_DATE, REG_USER, REG_UNO, 
				REG_AGENT, DNO
			) VALUES (
				t2.SNO, t2.JNO, t2.USER_KEY, t2.RECORD_DATE, t2.IN_RECOG_TIME, 
				t2.OUT_RECOG_TIME, t2.WORK_STATE, SYSDATE, 'TRG_IRIS_WORKER_DAILY_SET', t2.MOD_UNO, 
				t2.MOD_AGENT, t2.DNO
			)`

	query2 := `
		UPDATE IRIS_RECD_SET
		SET IS_DAILY_WORKER = 'Y'
		WHERE IRIS_NO = :1`

	for _, w := range worker {
		if _, err := tx.ExecContext(ctx, query, w.Sno, w.Jno, w.UserKey, w.RecordDate, w.InRecogTime, w.OutRecogTime, w.WorkState, agent, w.Dno); err != nil {
			return utils.CustomErrorf(err)
		}
		if _, err := tx.ExecContext(ctx, query2, w.IrisNo); err != nil {
			return utils.CustomErrorf(err)
		}
	}
	return nil
}

// 변경 이력 변경 전 데이터 조회
func (r *Repository) GetDailyWorkerBeforeList(ctx context.Context, db Queryer, workers entity.WorkerDailys) (entity.WorkerDailys, error) {
	var list entity.WorkerDailys
	query := `
		SELECT
			SNO, JNO, USER_KEY, RECORD_DATE, IN_RECOG_TIME,
			OUT_RECOG_TIME, IS_DEADLINE, WORK_STATE, IS_OVERTIME, WORK_HOUR,
			REG_DATE, REG_AGENT, REG_USER, REG_UNO
		FROM IRIS_WORKER_DAILY_SET
		WHERE USER_KEY = :1 
		AND TRUNC(RECORD_DATE) = TRUNC(:2)`

	for _, w := range workers {
		var dailyWorker entity.WorkerDaily
		if err := db.GetContext(ctx, &dailyWorker, query, w.UserKey, w.RecordDate); err != nil {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					dailyWorker.Sno = w.Sno
					dailyWorker.Jno = w.Jno
					dailyWorker.UserKey = w.UserKey
					dailyWorker.ReasonType = w.ReasonType
				} else {
					return entity.WorkerDailys{}, utils.CustomErrorf(err)
				}
			}
		}
		dailyWorker.ReasonType = w.ReasonType
		dailyWorker.Reason = w.Reason
		list = append(list, &dailyWorker)
	}
	return list, nil
}

// 변경 이력 변경 후 데이터 저장
func (r *Repository) AddHistoryDailyWorkers(ctx context.Context, tx Execer, workers entity.WorkerDailys) error {
	agent := utils.GetAgent()

	insetQuery := `
		INSERT INTO IRIS_WORKER_DAILY_HIS(
			SNO, JNO, USER_KEY, RECORD_DATE, IN_RECOG_TIME,
			OUT_RECOG_TIME, IS_DEADLINE, WORK_STATE, IS_OVERTIME, WORK_HOUR,
			HIS_STATUS, REASON, REASON_TYPE, REG_DATE, REG_AGENT,
			REG_USER, REG_UNO
		) VALUES (
			:1, :2, :3, :4, :5, 
			:6, :7, :8, :9, :10,
		    :11, :12, :13, :14, :15, 
		    :16, :17
		)`

	for _, w := range workers {
		if _, err := tx.ExecContext(ctx, insetQuery,
			w.Sno, w.Jno, w.UserKey, w.RecordDate, w.InRecogTime,
			w.OutRecogTime, w.IsDeadline, w.WorkState, w.IsOvertime, w.WorkHour,
			w.HisStatus, w.Reason, w.ReasonType, w.RegDate, agent,
			w.ModUser, w.ModUno,
		); err != nil {
			return utils.CustomErrorf(err)
		}
	}
	return nil
}

// 변경 이력 조회
func (r *Repository) GetHistoryDailyWorkers(ctx context.Context, db Queryer, startDate string, endDate string, sno int64, retry string, userKeys []string) (entity.WorkerDailys, error) {
	var list entity.WorkerDailys

	var columns []string
	columns = append(columns, "T1.REASON_TYPE")
	columns = append(columns, "T2.USER_NM")
	columns = append(columns, "T2.USER_ID")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	query := fmt.Sprintf(`
		SELECT
			HIS_STATUS,
			HIS_NAME,
			REASON_TYPE,
			REASON,
			REG_DATE,
			USER_ID,
			USER_NM,
			DEPARTMENT,
			JOB_NAME,
			FIXED_RECORD_DATE AS RECORD_DATE,
			FIXED_SNO AS SNO,
			IN_RECOG_TIME,
			OUT_RECOG_TIME,
			WORK_HOUR,
			WORK_STATE,
			IS_OVERTIME,
			IS_DEADLINE,
			CNO
		FROM (
			SELECT
				T1.HIS_STATUS,
				DECODE(T1.HIS_STATUS, 'AFTER', '후', 'BEFORE', '전', '') AS HIS_NAME,
				DECODE(T1.REASON_TYPE, 
						'01', '추가', 
						'02', '수정', 
						'03', '마감', 
						'04', '일괄공수입력', 
						'05', '프로젝트변경', 
						'06', '삭제', 
						'07', '마감취소', 
						'08', '수정/마감',
						'09', '엑셀업로드',  
						''
				) AS REASON_TYPE,
				T1.REASON,
				T1.REG_DATE,
				T2.USER_ID,
				T2.USER_NM,
				T2.DEPARTMENT,
				T3.JOB_NAME,
				COALESCE(
				  T1.RECORD_DATE,
				  MAX(T1.RECORD_DATE) OVER (
					PARTITION BY T1.USER_KEY, T1.REG_DATE
				  )
				) AS FIXED_RECORD_DATE,
				COALESCE(
				  T1.SNO,
				  MAX(T1.SNO) OVER (
					PARTITION BY T1.USER_KEY, T1.REG_DATE
				  )
				) AS FIXED_SNO,
				T1.IN_RECOG_TIME,
				T1.OUT_RECOG_TIME,
				T1.WORK_HOUR,
				DECODE(T1.WORK_STATE, '01', '출근', '02', '퇴근', '') AS WORK_STATE,
				T1.IS_OVERTIME,
				T1.IS_DEADLINE,
				T1.CNO
			FROM IRIS_WORKER_DAILY_HIS T1
			LEFT JOIN IRIS_WORKER_SET T2 ON T1.SNO = T2.SNO AND T1.USER_KEY = T2.USER_KEY
			LEFT JOIN S_JOB_INFO T3 ON T1.JNO = T3.JNO
			WHERE 1=1
			 AND ( ? = 1 OR T1.USER_KEY IN (?))
			%s
		)
		WHERE TO_CHAR(FIXED_RECORD_DATE, 'YYYY-MM-DD') BETWEEN ? AND ?
		  AND FIXED_SNO = ?
		ORDER BY
			REG_DATE DESC,
			USER_ID,
			FIXED_RECORD_DATE DESC,
			CASE WHEN HIS_STATUS = 'BEFORE' THEN 0 ELSE 1 END,
			HIS_STATUS DESC
		`, retryCondition)

	var args []any
	var err error
	if len(userKeys) > 0 { // userKeys가 있는 경우 userKeys에 해당하는 이력만 조회
		query, args, err = sqlx.In(query, 0, userKeys, startDate, endDate, sno)
	} else { // userKeys가 없는 경우 모든 이력 조회
		query, args, err = sqlx.In(query, 1, []string{"dummy"}, startDate, endDate, sno)
	}

	if err != nil {
		return nil, utils.CustomErrorf(err)
	}
	query = db.Rebind(query)

	if err = db.SelectContext(ctx, &list, query, args...); err != nil {
		return list, utils.CustomErrorf(err)
	}
	return list, nil
}

// 변경 이력 사유 조회
func (r *Repository) GetHistoryDailyWorkerReason(ctx context.Context, db Queryer, cno int64) (string, error) {
	var reason string

	query := `
		SELECT MAX(REASON)
		FROM IRIS_WORKER_DAILY_HIS
		WHERE (USER_KEY, TO_CHAR(REG_DATE, 'YYYY-MM-DD HH24:MI')) IN (
			SELECT USER_KEY, TO_CHAR(REG_DATE, 'YYYY-MM-DD HH24:MI')
			FROM IRIS_WORKER_DAILY_HIS
			WHERE cno = :1
		)`

	if err := db.GetContext(ctx, &reason, query, cno); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", utils.CustomErrorf(err)
	}
	return reason, nil
}
