package store

import (
	"context"
	"csm-api/entity"
	"csm-api/utils"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// func: 현장 프로젝트 조회
// @param
// - sno int64 현장 번호, targetDate time.Time: 현재시간
func (r *Repository) GetProjectList(ctx context.Context, db Queryer, sno int64, targetDate time.Time) (*entity.ProjectInfos, error) {
	projectInfos := entity.ProjectInfos{}

	// sno 변환: 0이면 NULL 처리, 아니면 Valid 값으로 설정
	var snoParam sql.NullInt64
	if sno != 0 {
		snoParam = sql.NullInt64{Valid: true, Int64: sno}
	} else {
		snoParam = sql.NullInt64{Valid: false}
	}

	formattedDate := targetDate.Format("2006-01-02")

	sql := `
			WITH base AS (
				SELECT
					t1.SNO,
					t1.JNO,
					t2.USER_ID,
					NVL(t2.USER_NM, ' ') AS USER_NM,
					TO_CHAR(t1.RECORD_DATE, 'YYYY-MM-DD') AS RECORD_DATE,
					NVL(t2.DEPARTMENT, ' ') AS DEPARTMENT,
					t2.WORKER_TYPE
				FROM IRIS_WORKER_DAILY_SET t1
				LEFT JOIN IRIS_WORKER_SET t2 ON t1.SNO = t2.SNO AND t1.JNO = t2.JNO AND t1.USER_KEY = t2.USER_KEY
				WHERE t1.SNO > 100
				AND (:1 IS NULL OR t1.SNO = :2) AND t1.RECORD_DATE < :3
			),
			worker_counts AS (
				SELECT
					SNO,
					JNO,
					COUNT(DISTINCT USER_ID || '-' || RECORD_DATE) AS WORKER_COUNT_ALL,
					COUNT(DISTINCT CASE 
									 WHEN RECORD_DATE = TO_CHAR(:4, 'YYYY-MM-DD')
									 THEN USER_ID || '-' || RECORD_DATE
									 END) AS WORKER_COUNT_DATE,
					COUNT(DISTINCT CASE 
									 WHEN RECORD_DATE = TO_CHAR(:5, 'YYYY-MM-DD')
									  AND (INSTR(DEPARTMENT, '하이테크') > 0 
									   OR INSTR(DEPARTMENT, 'HTENC') > 0 
									   OR INSTR(DEPARTMENT, 'HTE') > 0
									   OR WORKER_TYPE = '01')
									 THEN USER_ID || '-' || RECORD_DATE
									 END) AS WORKER_COUNT_HTENC,
					COUNT(DISTINCT CASE 
									 WHEN RECORD_DATE = TO_CHAR(:6, 'YYYY-MM-DD')
									  AND INSTR(DEPARTMENT, '하이테크') = 0
									  AND INSTR(DEPARTMENT, 'HTENC') = 0
									  AND INSTR(DEPARTMENT, 'HTE') = 0
									  AND WORKER_TYPE <> '01'
									  AND (INSTR(DEPARTMENT, '관리') > 0 
										   OR INSTR(USER_NM, '관리') > 0)
									 THEN USER_ID || '-' || RECORD_DATE
									 END) AS WORKER_COUNT_MANAGER,
					COUNT(DISTINCT CASE 
									 WHEN RECORD_DATE = TO_CHAR(:7, 'YYYY-MM-DD')
									  AND INSTR(DEPARTMENT, '하이테크') = 0
									  AND INSTR(DEPARTMENT, 'HTENC') = 0
									  AND INSTR(DEPARTMENT, 'HTE') = 0
									  AND WORKER_TYPE <> '01'
									  AND (INSTR(DEPARTMENT, '관리') = 0 
										   AND INSTR(USER_NM, '관리') = 0)
									 THEN USER_ID || '-' || RECORD_DATE
									 END) AS WORKER_COUNT_NOT_MANAGER
				FROM base
				GROUP BY SNO, JNO
			),
			equip AS (
				SELECT SNO, JNO, CNT
				FROM IRIS_EQUIP_TEMP
			),
			work_rate_info AS (
				SELECT WORK_RATE, IS_WORK_RATE, JNO FROM (
					SELECT R1.WORK_RATE, 'Y' AS IS_WORK_RATE, R1.JNO
					FROM IRIS_JOB_WORK_RATE R1
					WHERE R1.RECORD_DATE = TO_DATE(:8, 'YYYY-MM-DD')
					UNION ALL
					SELECT R3.WORK_RATE, 'N' AS IS_WORK_RATE, R3.JNO
					FROM IRIS_JOB_WORK_RATE R3
					JOIN (
						SELECT JNO, MAX(RECORD_DATE) AS MAX_RECORD_DATE
						FROM IRIS_JOB_WORK_RATE
						WHERE RECORD_DATE < TO_DATE(:9, 'YYYY-MM-DD')
						GROUP BY JNO
					) R4 ON R3.JNO = R4.JNO AND R3.RECORD_DATE = R4.MAX_RECORD_DATE
					WHERE NOT EXISTS (
						SELECT 1
						FROM IRIS_JOB_WORK_RATE R6
						WHERE R6.RECORD_DATE = TO_DATE(:10, 'YYYY-MM-DD')
						AND R6.JNO = R3.JNO
					)
					UNION ALL
					SELECT 0 AS WORK_RATE, 'N' AS IS_WORK_RATE, R7.JNO
					FROM (SELECT DISTINCT JNO FROM IRIS_JOB_WORK_RATE) R7
					WHERE NOT EXISTS (
						SELECT 1
						FROM IRIS_JOB_WORK_RATE R8
						WHERE R8.JNO = R7.JNO
						AND R8.RECORD_DATE <= TO_DATE(:11, 'YYYY-MM-DD')
					)
				)
			)
			SELECT
				t1.SNO,
				t1.JNO,
				t1.IS_USE,
				t1.IS_DEFAULT,
				t1.REG_DATE,
				t1.REG_USER,
				t1.REG_UNO,
				t1.MOD_DATE,
				t1.MOD_USER,
				t1.MOD_UNO,
				NVL(t8.WORK_RATE, 0) AS WORK_RATE,
				NVL(t8.IS_WORK_RATE, 'N') AS IS_WORK_RATE,
				t2.JOB_TYPE AS PROJECT_TYPE,
				t6.CD_NM AS PROJECT_TYPE_NM,
				t2.JOB_NO AS PROJECT_NO,
				t2.JOB_NAME AS PROJECT_NM,
				t2.JOB_YEAR AS PROJECT_YEAR,
				t2.JOB_LOC AS PROJECT_LOC,
				t2.JOB_CODE AS PROJECT_CODE,
				t4.KIND_NAME AS PROJECT_CODE_NAME,
				t7.CANCEL_DAY,
				t3.SITE_NM,
				t2.COMP_CODE,
				t2.COMP_NICK,
				t2.COMP_NAME,
				t2.COMP_ETC,
				t2.ORDER_COMP_CODE,
				t2.ORDER_COMP_NICK,
				t2.ORDER_COMP_NAME,
				t2.ORDER_COMP_JOB_NAME,
				t2.JOB_LOC_NAME AS PROJECT_LOC_NAME,
				t2.JOB_PM,
				t2.JOB_PM_NAME,
				t2.JOB_PE,
				TO_DATE(t2.JOB_SD, 'YYYY-MM-DD') AS PROJECT_STDT,
				TO_DATE(t2.JOB_ED, 'YYYY-MM-DD') AS PROJECT_EDDT,
				TO_DATE(t2.REG_DATE, 'YYYY-MM-DD HH24:MI:SS') AS PROJECT_REG_DATE,
				TO_DATE(t2.MOD_DATE, 'YYYY-MM-DD HH24:MI:SS') AS PROJECT_MOD_DATE,
				t2.JOB_STATE AS PROJECT_STATE,
				t5.CD_NM AS PROJECT_STATE_NM,
				t2.MOC_NO,
				t2.WO_NO,
				NVL(wc.WORKER_COUNT_ALL, 0) AS WORKER_COUNT_ALL,
				NVL(wc.WORKER_COUNT_DATE, 0) AS WORKER_COUNT_DATE,
				NVL(wc.WORKER_COUNT_HTENC, 0) AS WORKER_COUNT_HTENC,
				NVL(wc.WORKER_COUNT_MANAGER, 0) AS WORKER_COUNT_MANAGER,
				NVL(wc.WORKER_COUNT_NOT_MANAGER, 0) AS WORKER_COUNT_NOT_MANAGER,
				NVL(eq.CNT, 0) AS EQUIP_COUNT
			FROM IRIS_SITE_JOB t1
			INNER JOIN S_JOB_INFO t2 ON t1.JNO = t2.JNO
			INNER JOIN IRIS_SITE_SET t3 ON t1.SNO = t3.SNO
			INNER JOIN TIMESHEET.JOB_KIND_CODE t4 ON t2.JOB_CODE = t4.KIND_CODE
			INNER JOIN TIMESHEET.SYS_CODE_SET t5 ON t5.MINOR_CD = t2.job_state AND t5.major_cd = 'JOB_STATE'
			INNER JOIN TIMESHEET.SYS_CODE_SET t6 ON t6.MINOR_CD = t2.JOB_TYPE AND t6.major_cd = 'JOB_TYPE' 
			INNER JOIN ( SELECT J.JNO, C.UDF_VAL_03 AS CANCEL_DAY FROM IRIS_JOB_SET J INNER JOIN IRIS_CODE_SET C ON J.CANCEL_CODE =  C.CODE ) t7 ON t1.JNO = t7.JNO
			LEFT JOIN work_rate_info t8 ON t1.JNO = t8.JNO
			LEFT JOIN worker_counts wc ON t1.SNO = wc.SNO AND t1.JNO = wc.JNO
			LEFT JOIN equip eq ON t1.SNO = eq.SNO  AND t1.JNO = eq.JNO
			WHERE t1.SNO > 100
			AND (:12 IS NULL OR t1.SNO = :13)
			ORDER BY IS_DEFAULT DESC, JNO DESC`

	if err := db.SelectContext(ctx, &projectInfos, sql,
		snoParam, snoParam, targetDate, targetDate, targetDate,
		targetDate, targetDate, formattedDate, formattedDate, formattedDate,
		formattedDate, snoParam, snoParam,
	); err != nil {
		return &projectInfos, utils.CustomErrorf(err)
	}

	return &projectInfos, nil
}

// func: 프로젝트 근로자 수 조회
// @param
// - sno int64 현장 번호, targetDate time.Time: 현재시간
func (r *Repository) GetProjectWorkerCountList(ctx context.Context, db Queryer, targetDate time.Time) (*entity.ProjectInfos, error) {
	list := entity.ProjectInfos{}

	query := `
				WITH base AS (
					SELECT
						t1.SNO,
						t1.JNO,
						t2.USER_ID,
						NVL(t2.USER_NM, ' ') AS USER_NM,
						TO_CHAR(t1.RECORD_DATE, 'YYYY-MM-DD') AS RECORD_DATE,
						NVL(t2.DEPARTMENT, ' ') AS DEPARTMENT,
						t2.WORKER_TYPE
					FROM IRIS_WORKER_DAILY_SET t1
					LEFT JOIN IRIS_WORKER_SET t2 ON t1.SNO = t2.SNO AND t1.JNO = t2.JNO AND t1.USER_KEY = t2.USER_KEY
					WHERE t1.SNO > 100
				),
				worker_counts AS (
					SELECT
						SNO,
						JNO,
						COUNT(DISTINCT USER_ID || '-' || RECORD_DATE) AS WORKER_COUNT_ALL,
						COUNT(DISTINCT CASE 
										 WHEN RECORD_DATE = TO_CHAR(:1, 'YYYY-MM-DD')
										 THEN USER_ID || '-' || RECORD_DATE
										 END) AS WORKER_COUNT_DATE,
						COUNT(DISTINCT CASE 
										 WHEN RECORD_DATE = TO_CHAR(:2, 'YYYY-MM-DD')
										  AND (INSTR(DEPARTMENT, '하이테크') > 0 
										   OR INSTR(DEPARTMENT, 'HTENC') > 0 
										   OR INSTR(DEPARTMENT, 'HTE') > 0
										   OR WORKER_TYPE = '01')
										 THEN USER_ID || '-' || RECORD_DATE
										 END) AS WORKER_COUNT_HTENC,
						COUNT(DISTINCT CASE 
										 WHEN RECORD_DATE = TO_CHAR(:3, 'YYYY-MM-DD')
										  AND INSTR(DEPARTMENT, '하이테크') = 0
										  AND INSTR(DEPARTMENT, 'HTENC') = 0
										  AND INSTR(DEPARTMENT, 'HTE') = 0
										  AND WORKER_TYPE <> '01'
										  AND (INSTR(DEPARTMENT, '관리') > 0 
											   OR INSTR(USER_NM, '관리') > 0)
										 THEN USER_ID || '-' || RECORD_DATE
										 END) AS WORKER_COUNT_MANAGER,
						COUNT(DISTINCT CASE 
										 WHEN RECORD_DATE = TO_CHAR(:4, 'YYYY-MM-DD')
										  AND INSTR(DEPARTMENT, '하이테크') = 0
										  AND INSTR(DEPARTMENT, 'HTENC') = 0
										  AND INSTR(DEPARTMENT, 'HTE') = 0
										  AND WORKER_TYPE <> '01'
										  AND (INSTR(DEPARTMENT, '관리') = 0 
											   AND INSTR(USER_NM, '관리') = 0)
										 THEN USER_ID || '-' || RECORD_DATE
										 END) AS WORKER_COUNT_NOT_MANAGER
					FROM base
					GROUP BY SNO, JNO
				),
				equip AS (
					SELECT SNO, JNO, CNT
					FROM IRIS_EQUIP_TEMP
				)
				SELECT 
					t1.SNO,
					t1.JNO,
					NVL(wc.WORKER_COUNT_ALL, 0) AS WORKER_COUNT_ALL,
					NVL(wc.WORKER_COUNT_DATE, 0) AS WORKER_COUNT_DATE,
					NVL(wc.WORKER_COUNT_HTENC, 0) AS WORKER_COUNT_HTENC,
					NVL(wc.WORKER_COUNT_MANAGER, 0) AS WORKER_COUNT_MANAGER,
					NVL(wc.WORKER_COUNT_NOT_MANAGER, 0) AS WORKER_COUNT_NOT_MANAGER,
					NVL(eq.CNT, 0) AS EQUIP_COUNT
				FROM IRIS_SITE_JOB t1
				LEFT JOIN worker_counts wc ON t1.SNO = wc.SNO AND t1.JNO = wc.JNO
				LEFT JOIN equip eq ON t1.SNO = eq.SNO  AND t1.JNO = eq.JNO
				WHERE t1.SNO > 100`

	if err := db.SelectContext(ctx, &list, query, targetDate, targetDate, targetDate, targetDate); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return &list, nil
}

// func: 프로젝트별 출근 안전관리자 수
// @param
// - targetDate: 현재시간
func (r *Repository) GetProjectSafeWorkerCountList(ctx context.Context, db Queryer, targetDate time.Time) (*entity.ProjectSafeCounts, error) {
	list := entity.ProjectSafeCounts{}

	query := `
				WITH htenc_cnt AS (
					SELECT T1.SNO, T1.JNO, T2.USER_NM
					FROM IRIS_WORKER_DAILY_SET t1
					LEFT JOIN IRIS_WORKER_SET t2 ON t1.SNO = t2.SNO AND t1.JNO = t2.JNO AND t1.USER_KEY = t2.USER_KEY
					WHERE TO_CHAR(T1.RECORD_DATE, 'YYYY-MM-DD') = TO_CHAR(:1, 'YYYY-MM-DD')
					AND (
						INSTR(T2.DEPARTMENT, '하이테크') > 0 
						OR INSTR(T2.DEPARTMENT, 'HTENC') > 0 
						OR INSTR(T2.DEPARTMENT, 'HTE') > 0
						OR T2.WORKER_TYPE = '01'
					)
				),
				safe_cnt AS (
				  SELECT t1.JNO, USER_NAME
				  FROM JOB_MANAGER t1
				  JOIN S_SYS_USER_SET t2 ON t2.UNO = t1.UNO
				  WHERE t1.AUTH = 'SAFETY_MANAGER'
				  UNION
				  SELECT t1.JNO, USER_NAME
				  FROM S_JOB_MEMBER_LIST t1
				  JOIN S_SYS_USER_SET t2 ON t2.UNO = t1.UNO
				  WHERE t1.COMP_TYPE = 'H'
					AND t1.FUNC_CODE = 510
					AND t1.CHARGE = '21'
					AND t1.IS_USE = 'Y'
				),
				cnt AS (
					SELECT 
					  ht.SNO,
					  ht.JNO,
					  COUNT(*) AS SAFE_COUNT
					FROM htenc_cnt ht
					WHERE EXISTS (
					  SELECT 1
					  FROM safe_cnt s
					  WHERE s.JNO = ht.JNO
						AND INSTR(ht.USER_NM, s.USER_NAME) > 0
					)
					GROUP BY ht.SNO, ht.JNO
				)
				SELECT 
					t1.SNO, 
					t1.JNO, 
					NVL(t2.SAFE_COUNT, 0) AS SAFE_COUNT
				FROM IRIS_SITE_JOB t1
				LEFT JOIN cnt t2 ON t1.SNO = t2.SNO AND t1.JNO = t2.JNO`

	if err := db.SelectContext(ctx, &list, query, targetDate); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return &list, nil
}

// func: 프로젝트 조회(이름)
// @param
// -
func (r *Repository) GetProjectNmList(ctx context.Context, db Queryer, role int, uno int64) (*entity.ProjectInfos, error) {
	projectInfos := entity.ProjectInfos{}

	sql := `
		WITH USER_IN_JNO AS (
			SELECT JNO
			FROM S_JOB_MEMBER_LIST
			WHERE 1 = :1 OR UNO = :2
		UNION
			SELECT JNO
			FROM JOB_SUBCON_INFO
			WHERE ID = :3
	)
			SELECT
    			t1.SNO,
				t1.JNO,
				t2.JOB_NAME as PROJECT_NM,
				t1.WORK_RATE,
				t5.CANCEL_DAY
			FROM
				IRIS_SITE_JOB t1
				INNER JOIN S_JOB_INFO t2 ON t1.JNO = t2.JNO AND t2.JNO IN (SELECT * FROM USER_IN_JNO)
				INNER JOIN IRIS_SITE_SET t3 ON t1.SNO = t3.SNO
				INNER JOIN TIMESHEET.JOB_KIND_CODE t4 ON t2.JOB_CODE = t4.KIND_CODE
				INNER JOIN ( SELECT J.JNO, C.UDF_VAL_03 AS CANCEL_DAY FROM IRIS_JOB_SET J INNER JOIN IRIS_CODE_SET C ON J.CANCEL_CODE =  C.CODE ) t5 ON t1.JNO = t5.JNO
			WHERE t1.sno > 100
			AND t1.IS_USE = 'Y'
			ORDER BY
				t1.IS_DEFAULT DESC, JNO DESC`
	if err := db.SelectContext(ctx, &projectInfos, sql, role, uno, uno); err != nil {
		return &projectInfos, utils.CustomErrorf(err)
	}

	return &projectInfos, nil
}

// func: 공사관리시스템 등록 프로젝트 전체 조회
// @param
// -
func (r *Repository) GetUsedProjectList(ctx context.Context, db Queryer, pageSql entity.PageSql, search entity.JobInfo, retry string, includeJno string, snoString string) (*entity.JobInfos, error) {
	list := entity.JobInfos{}

	condition := ""
	condition = utils.StringWhereConvert(condition, search.JobNo.NullString, "t2.JOB_NO")
	condition = utils.StringWhereConvert(condition, search.CompName.NullString, "t2.COMP_NAME")
	condition = utils.StringWhereConvert(condition, search.OrderCompName.NullString, "t2.ORDER_COMP_NAME")
	condition = utils.StringWhereConvert(condition, search.JobName.NullString, "t2.JOB_NAME")
	condition = utils.StringWhereConvert(condition, search.JobPmName.NullString, "t2.JOB_PM_NAME")
	condition = utils.StringWhereConvert(condition, search.JobSd.NullString, "t2.JOB_SD")
	condition = utils.StringWhereConvert(condition, search.JobEd.NullString, "t2.JOB_ED")
	condition = utils.StringWhereConvert(condition, search.CdNm.NullString, "t5.CD_NM")

	var columns []string
	columns = append(columns, "t1.JNO")
	columns = append(columns, "t2.JOB_NO")
	columns = append(columns, "t2.JOB_NAME")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	var order string
	if pageSql.Order.Valid {
		order = pageSql.Order.String
	} else {
		order = "JNO DESC, JOB_NO ASC"
	}

	var jnoCondition string
	if includeJno != "undefined" && includeJno != "" {
		parseInt, _ := strconv.ParseInt(includeJno, 10, 64)
		jnoCondition = fmt.Sprintf(`
			AND t1.SNO = (
				SELECT SNO
				FROM IRIS_SITE_JOB
				WHERE JNO = %d
			)`, parseInt)
	}

	var snoCondition string
	if snoString != "undefined" && snoString != "" {
		parseInt, _ := strconv.ParseInt(snoString, 10, 64)
		snoCondition = fmt.Sprintf(`AND t1.SNO = %d`, parseInt)
	}

	query := fmt.Sprintf(`
				SELECT *
				FROM (
					SELECT ROWNUM AS RNUM, sorted_data.*
					FROM (
						SELECT 
							t1.JNO,
							t1.SNO,
							t2.JOB_NAME,
							t2.JOB_NO,
							t2.JOB_SD,
							t2.JOB_ED,
							t2.COMP_NAME,
							t2.ORDER_COMP_NAME,
							t2.JOB_PM_NAME,
							t5.CD_NM
						FROM
							IRIS_SITE_JOB t1
							INNER JOIN S_JOB_INFO t2 ON t1.JNO = t2.JNO
							INNER JOIN IRIS_SITE_SET t3 ON t1.SNO = t3.SNO
							INNER JOIN TIMESHEET.JOB_KIND_CODE t4 ON t2.JOB_CODE = t4.KIND_CODE
							INNER JOIN TIMESHEET.SYS_CODE_SET t5 ON t5.MINOR_CD = t2.job_state AND t5.major_cd = 'JOB_STATE'
						WHERE t1.SNO > 100
						AND t1.IS_USE = 'Y'
						%s %s %s %s
						ORDER BY %s
					) sorted_data
					WHERE ROWNUM <= :1
				)
				WHERE RNUM > :2`, jnoCondition, snoCondition, condition, retryCondition, order)

	if err := db.SelectContext(ctx, &list, query, pageSql.EndNum, pageSql.StartNum); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return &list, nil
}

// func: 공사관리시스템 등록 프로젝트 전체 조회 개수
// @param
// -
func (r *Repository) GetUsedProjectCount(ctx context.Context, db Queryer, search entity.JobInfo, retry string, includeJno string, snoString string) (int, error) {
	var count int

	condition := ""
	condition = utils.StringWhereConvert(condition, search.JobNo.NullString, "t2.JOB_NO")
	condition = utils.StringWhereConvert(condition, search.CompName.NullString, "t2.COMP_NAME")
	condition = utils.StringWhereConvert(condition, search.OrderCompName.NullString, "t2.ORDER_COMP_NAME")
	condition = utils.StringWhereConvert(condition, search.JobName.NullString, "t2.JOB_NAME")
	condition = utils.StringWhereConvert(condition, search.JobPmName.NullString, "t2.JOB_PM_NAME")
	condition = utils.StringWhereConvert(condition, search.JobSd.NullString, "t2.JOB_SD")
	condition = utils.StringWhereConvert(condition, search.JobEd.NullString, "t2.JOB_ED")
	condition = utils.StringWhereConvert(condition, search.CdNm.NullString, "t5.CD_NM")

	var columns []string
	columns = append(columns, "t1.JNO")
	columns = append(columns, "t2.JOB_NO")
	columns = append(columns, "t2.JOB_NAME")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	var jnoCondition string
	if includeJno != "undefined" && includeJno != "" {
		parseInt, _ := strconv.ParseInt(includeJno, 10, 64)
		jnoCondition = fmt.Sprintf(`
			AND t1.SNO = (
				SELECT SNO
				FROM IRIS_SITE_JOB
				WHERE JNO = %d
			)`, parseInt)
	}

	var snoCondition string
	if snoString != "undefined" && snoString != "" {
		parseInt, _ := strconv.ParseInt(snoString, 10, 64)
		snoCondition = fmt.Sprintf(`AND t1.SNO = %d`, parseInt)
	}

	query := fmt.Sprintf(`
				SELECT 
					COUNT(*)
				FROM
					IRIS_SITE_JOB t1
					INNER JOIN S_JOB_INFO t2 ON t1.JNO = t2.JNO
					INNER JOIN IRIS_SITE_SET t3 ON t1.SNO = t3.SNO
					INNER JOIN TIMESHEET.JOB_KIND_CODE t4 ON t2.JOB_CODE = t4.KIND_CODE
					INNER JOIN TIMESHEET.SYS_CODE_SET t5 ON t5.MINOR_CD = t2.job_state AND t5.major_cd = 'JOB_STATE'
				WHERE t1.SNO > 100
				AND t1.IS_USE = 'Y'
				%s %s %s %s`, jnoCondition, snoCondition, condition, retryCondition)

	if err := db.GetContext(ctx, &count, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, utils.CustomErrorf(err)
	}

	return count, nil
}

// func: 프로젝트 전체 조회
// @param
// -
func (r *Repository) GetAllProjectList(ctx context.Context, db Queryer, pageSql entity.PageSql, search entity.JobInfo, isAll int, retry string) (*entity.JobInfos, error) {
	list := entity.JobInfos{}

	condition := "1 = 1"
	condition = utils.StringWhereConvert(condition, search.JobNo.NullString, "J.JOB_NO")
	condition = utils.StringWhereConvert(condition, search.CompName.NullString, "J.COMP_NAME")
	condition = utils.StringWhereConvert(condition, search.OrderCompName.NullString, "J.ORDER_COMP_NAME")
	condition = utils.StringWhereConvert(condition, search.JobName.NullString, "J.JOB_NAME")
	condition = utils.StringWhereConvert(condition, search.JobPmName.NullString, "J.JOB_PM_NAME")
	condition = utils.StringWhereConvert(condition, search.JobSd.NullString, "J.JOB_SD")
	condition = utils.StringWhereConvert(condition, search.JobEd.NullString, "J.JOB_ED")
	condition = utils.StringWhereConvert(condition, search.CdNm.NullString, "SC.CD_NM")

	var columns []string
	columns = append(columns, "J.JNO")
	columns = append(columns, "J.JOB_NO")
	columns = append(columns, "J.JOB_NAME")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	var order string
	if pageSql.Order.Valid {
		order = pageSql.Order.String
	} else {
		order = "JNO DESC"
	}

	query := fmt.Sprintf(`
				SELECT
					0 AS RNUM,
				    0 AS JNO,
					100 AS SNO,
					'전체' AS JOB_NAME,
					'' AS JOB_NO,
					'-' AS JOB_SD,
					'-' AS JOB_ED,
					'-' AS COMP_NAME,
					'-' AS ORDER_COMP_NAME,
					'-' AS JOB_PM_NAME,
					'-' AS CD_NM
				FROM DUAL
				WHERE 1=:1
			UNION
				SELECT *
				FROM (
					SELECT ROWNUM AS RNUM, sorted_data.*
					FROM (
						SELECT 
							J.JNO,
							S.SNO, 
							J.JOB_NAME, 
							J.JOB_NO, 
							J.JOB_SD, 
							J.JOB_ED, 
							J.COMP_NAME, 
							J.ORDER_COMP_NAME, 
							J.JOB_PM_NAME, 
							SC.CD_NM 
						FROM 
							S_JOB_INFO J 
						INNER JOIN 
							TIMESHEET.job_kind_code JC 
						ON 
							J.job_code = JC.kind_code 
						LEFT JOIN 
							IRIS_SITE_JOB S
						ON
							S.JNO = J.JNO
						INNER JOIN 
							TIMESHEET.SYS_CODE_SET SC 
						ON 
							J.job_state = SC.minor_cd 
							AND SC.MAJOR_CD = 'JOB_STATE' 
						WHERE %s %s
						ORDER BY %s
					) sorted_data
					WHERE ROWNUM <= :2
				)
				WHERE RNUM > :3`, condition, retryCondition, order)

	if err := db.SelectContext(ctx, &list, query, isAll, pageSql.EndNum, pageSql.StartNum); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return &list, nil
}

// func: 프로젝트 개수 조회 개수
// @param
// -
func (r *Repository) GetAllProjectCount(ctx context.Context, db Queryer, search entity.JobInfo, retry string) (int, error) {
	var count int

	condition := "1 = 1"
	condition = utils.StringWhereConvert(condition, search.JobNo.NullString, "J.JOB_NO")
	condition = utils.StringWhereConvert(condition, search.CompName.NullString, "J.COMP_NAME")
	condition = utils.StringWhereConvert(condition, search.OrderCompName.NullString, "J.ORDER_COMP_NAME")
	condition = utils.StringWhereConvert(condition, search.JobName.NullString, "J.JOB_NAME")
	condition = utils.StringWhereConvert(condition, search.JobPmName.NullString, "J.JOB_PM_NAME")
	condition = utils.StringWhereConvert(condition, search.JobSd.NullString, "J.JOB_SD")
	condition = utils.StringWhereConvert(condition, search.JobEd.NullString, "J.JOB_ED")
	condition = utils.StringWhereConvert(condition, search.CdNm.NullString, "SC.CD_NM")

	var columns []string
	columns = append(columns, "J.JNO")
	columns = append(columns, "J.JOB_NO")
	columns = append(columns, "J.JOB_PM_NAME")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	query := fmt.Sprintf(`
				SELECT 
					count(*)
				FROM 
					S_JOB_INFO J 
				INNER JOIN 
					TIMESHEET.job_kind_code JC 
				ON 
					J.job_code = JC.kind_code 
				INNER JOIN 
					TIMESHEET.SYS_CODE_SET SC 
				ON 
					J.job_state = SC.minor_cd 
					AND SC.MAJOR_CD = 'JOB_STATE' 
				WHERE %s %s`, condition, retryCondition)

	if err := db.GetContext(ctx, &count, query); err != nil {
		return 0, utils.CustomErrorf(err)
	}

	return count, nil
}

// func: 본인이 속한 프로젝트 조회
// @param
// - UNO
func (r *Repository) GetStaffProjectList(ctx context.Context, db Queryer, pageSql entity.PageSql, searchSql entity.JobInfo, uno sql.NullInt64, retry string) (*entity.JobInfos, error) {

	list := entity.JobInfos{}

	condition := "1=1"
	condition = utils.StringWhereConvert(condition, searchSql.JobNo.NullString, "J.JOB_NO")
	condition = utils.StringWhereConvert(condition, searchSql.CompName.NullString, "J.COMP_NAME")
	condition = utils.StringWhereConvert(condition, searchSql.OrderCompName.NullString, "J.ORDER_COMP_NAME")
	condition = utils.StringWhereConvert(condition, searchSql.JobName.NullString, "J.JOB_NAME")
	condition = utils.StringWhereConvert(condition, searchSql.JobPmName.NullString, "J.JOB_PM_NAME")
	condition = utils.StringWhereConvert(condition, searchSql.JobSd.NullString, "J.JOB_SD")
	condition = utils.StringWhereConvert(condition, searchSql.JobEd.NullString, "J.JOB_ED")
	condition = utils.StringWhereConvert(condition, searchSql.CdNm.NullString, "SC.CD_NM")

	var columns []string
	columns = append(columns, "J.JNO")
	columns = append(columns, "J.JOB_NO")
	columns = append(columns, "J.JOB_NAME")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	var order string
	if pageSql.Order.Valid {
		order = pageSql.Order.String
	} else {
		order = "JNO DESC"
	}

	query := fmt.Sprintf(`
		SELECT *
		FROM (
			SELECT ROWNUM AS RNUM, sorted_data.*
				FROM (
					SELECT 
						J.JNO,
						S.SNO,
						J.JOB_NAME, 
						J.JOB_NO, 
						J.JOB_SD, 
						J.JOB_ED, 
						J.COMP_NAME, 
						J.ORDER_COMP_NAME, 
						J.JOB_PM_NAME, 
						SC.CD_NM 
					FROM 
						S_JOB_INFO J 
					INNER JOIN 
						(SELECT * FROM S_JOB_MEMBER_LIST WHERE UNO = :1) JM 
					ON 
						J.JNO = JM.JNO
					LEFT JOIN
						IRIS_SITE_JOB S
					ON
						S.JNO = J.JNO
					INNER JOIN 
						TIMESHEET.SYS_CODE_SET SC 
					ON 
						J.job_state = SC.minor_cd 
						AND SC.MAJOR_CD = 'JOB_STATE'
					WHERE %s %s
					ORDER BY %s
					) sorted_data
				WHERE ROWNUM <= :2
			) 
			WHERE RNUM > :3`, condition, retryCondition, order)

	if err := db.SelectContext(ctx, &list, query, uno, pageSql.EndNum, pageSql.StartNum); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return &list, nil
}

// func: 본인이 속한 프로젝트 개수
// @param
// - UNO
func (r *Repository) GetStaffProjectCount(ctx context.Context, db Queryer, searchSql entity.JobInfo, uno sql.NullInt64, retry string) (int, error) {
	var count int

	condition := "1=1"
	condition = utils.StringWhereConvert(condition, searchSql.JobNo.NullString, "J.JOB_NO")
	condition = utils.StringWhereConvert(condition, searchSql.CompName.NullString, "J.COMP_NAME")
	condition = utils.StringWhereConvert(condition, searchSql.OrderCompName.NullString, "J.ORDER_COMP_NAME")
	condition = utils.StringWhereConvert(condition, searchSql.JobName.NullString, "J.JOB_NAME")
	condition = utils.StringWhereConvert(condition, searchSql.JobPmName.NullString, "J.JOB_PM_NAME")
	condition = utils.StringWhereConvert(condition, searchSql.JobSd.NullString, "J.JOB_SD")
	condition = utils.StringWhereConvert(condition, searchSql.JobEd.NullString, "J.JOB_ED")
	condition = utils.StringWhereConvert(condition, searchSql.CdNm.NullString, "SC.CD_NM")

	var columns []string
	columns = append(columns, "J.JNO")
	columns = append(columns, "J.JOB_NO")
	columns = append(columns, "J.JOB_NAME")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	query := fmt.Sprintf(`
				SELECT 
					COUNT(*)
				FROM S_JOB_INFO J 
				INNER JOIN 
					(SELECT * FROM S_JOB_MEMBER_LIST WHERE UNO = :1) JM 
				ON 
					J.JNO = JM.JNO 
				INNER JOIN 
					TIMESHEET.SYS_CODE_SET SC 
				ON 
					J.job_state = SC.minor_cd 
					AND SC.MAJOR_CD = 'JOB_STATE'
				WHERE %s %s`, condition, retryCondition)

	if err := db.GetContext(ctx, &count, query, uno); err != nil {
		return 0, utils.CustomErrorf(err)
	}

	return count, nil
}

// func: 본인이 속한 프로젝트 이름 조회
// @param
// - uno
func (r *Repository) GetProjectNmUnoList(ctx context.Context, db Queryer, uno sql.NullInt64, role int) (*entity.ProjectInfos, error) {
	projectInfos := entity.ProjectInfos{}

	query := `SELECT 
    			JNO, 
    			JOB_NAME as PROJECT_NM 
			  FROM 
			    S_JOB_INFO 
			  WHERE
			      JOB_STATE = 'Y' AND
			      1=:1 OR 
			      JNO IN (SELECT DISTINCT(JNO) 
						  FROM S_JOB_MEMBER_LIST 
						  WHERE UNO = :2)
			  ORDER BY 
			      JNO DESC`

	if err := db.SelectContext(ctx, &projectInfos, query, role, uno); err != nil {
		return &projectInfos, utils.CustomErrorf(err)
	}

	return &projectInfos, nil
}

// func: 현장근태에 등록되지 않은 프로젝트
// @param
// -
func (r *Repository) GetNonUsedProjectList(ctx context.Context, db Queryer, page entity.PageSql, search entity.NonUsedProject, retry string) (*entity.NonUsedProjects, error) {
	nonProjects := entity.NonUsedProjects{}

	condition := ""

	condition = utils.Int64WhereConvert(condition, search.Jno.NullInt64, "t1.JNO")
	condition = utils.StringWhereConvert(condition, search.JobNo.NullString, "t1.JOB_NO")
	condition = utils.StringWhereConvert(condition, search.JobName.NullString, "t1.JOB_NAME")
	condition = utils.Int64WhereConvert(condition, search.JobYear.NullInt64, "t1.JOB_YEAR")
	condition = utils.StringWhereConvert(condition, search.JobSd.NullString, "t1.JOB_SD")
	condition = utils.StringWhereConvert(condition, search.JobEd.NullString, "t1.JOB_ED")
	condition = utils.StringWhereConvert(condition, search.JobPmNm.NullString, "t2.USER_NAME")

	var columns []string
	columns = append(columns, "t1.JNO")
	columns = append(columns, "t1.JOB_NO")
	columns = append(columns, "t1.JOB_NAME")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	var order string
	if page.Order.Valid {
		order = page.Order.String
	} else {
		order = `JNO DESC,
				CASE 
					WHEN t1.REG_DATE IS NULL THEN t1.MOD_DATE 
					WHEN t1.MOD_DATE IS NULL THEN t1.REG_DATE 
					ELSE GREATEST(t1.REG_DATE, t1.MOD_DATE) 
				END DESC NULLS LAST`
	}

	query := fmt.Sprintf(`
								SELECT *
								FROM (
									SELECT ROWNUM AS RNUM, sorted_data.*
									FROM (
									    SELECT 
											t1.JNO,
											t1.JOB_NO,
											t1.JOB_NAME,
											t1.JOB_YEAR,
											t1.JOB_SD,
											t1.JOB_ED,
											t1.COMP_NAME,
											t1.ORDER_COMP_NAME,
											t2.USER_NAME AS JOB_PM_NM,
											t2.DUTY_NAME,
											t5.CD_NM
										FROM s_job_info t1
										LEFT JOIN COMMON.V_BIZ_USER_INFO t2 ON t1.JOB_PM = t2.UNO
										LEFT JOIN TIMESHEET.SYS_CODE_SET t5 ON t5.MINOR_CD = t1.job_state AND t5.major_cd = 'JOB_STATE'
										WHERE t1.JOB_STATE = 'Y'
										AND t1.JNO NOT IN(
											SELECT JNO
											FROM IRIS_SITE_JOB
										)
										%s %s
										ORDER BY %s 
									) sorted_data
									WHERE ROWNUM <= :1
								)
								WHERE RNUM > :2`, condition, retryCondition, order)

	if err := db.SelectContext(ctx, &nonProjects, query, page.EndNum, page.StartNum); err != nil {
		return &nonProjects, utils.CustomErrorf(err)
	}

	return &nonProjects, nil
}

// func: 현장근태에 등록되지 않은 프로젝트 수
// @param
// -
func (r *Repository) GetNonUsedProjectCount(ctx context.Context, db Queryer, search entity.NonUsedProject, retry string) (int, error) {
	var count int

	condition := ""

	condition = utils.Int64WhereConvert(condition, search.Jno.NullInt64, "t1.JNO")
	condition = utils.StringWhereConvert(condition, search.JobNo.NullString, "t1.JOB_NO")
	condition = utils.StringWhereConvert(condition, search.JobName.NullString, "t1.JOB_NAME")
	condition = utils.Int64WhereConvert(condition, search.JobYear.NullInt64, "t1.JOB_YEAR")
	condition = utils.StringWhereConvert(condition, search.JobSd.NullString, "t1.JOB_SD")
	condition = utils.StringWhereConvert(condition, search.JobEd.NullString, "t1.JOB_ED")
	condition = utils.StringWhereConvert(condition, search.JobPmNm.NullString, "t2.USER_NAME")

	var columns []string
	columns = append(columns, "t1.JNO")
	columns = append(columns, "t1.JOB_NO")
	columns = append(columns, "t1.JOB_NAME")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	query := fmt.Sprintf(`
								SELECT 
									COUNT(*)
								FROM s_job_info t1
								LEFT JOIN COMMON.V_BIZ_USER_INFO t2 ON t1.JOB_PM = t2.UNO
								WHERE t1.JOB_STATE = 'Y'
								AND t1.JNO NOT IN(
									SELECT JNO
									FROM IRIS_SITE_JOB
								)
								%s %s`, condition, retryCondition)

	if err := db.GetContext(ctx, &count, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, utils.CustomErrorf(err)
	}

	return count, nil
}

// func: 현장근태에 등록되지 않은 프로젝트(타입별)
// @param
// -
func (r *Repository) GetNonUsedProjectListByType(ctx context.Context, db Queryer, page entity.PageSql, search entity.NonUsedProject, retry string, typeString string) (*entity.NonUsedProjects, error) {
	nonProjects := entity.NonUsedProjects{}

	condition := ""

	condition = utils.Int64WhereConvert(condition, search.Jno.NullInt64, "t1.JNO")
	condition = utils.StringWhereConvert(condition, search.JobNo.NullString, "t1.JOB_NO")
	condition = utils.StringWhereConvert(condition, search.JobName.NullString, "t1.JOB_NAME")
	condition = utils.Int64WhereConvert(condition, search.JobYear.NullInt64, "t1.JOB_YEAR")
	condition = utils.StringWhereConvert(condition, search.JobSd.NullString, "t1.JOB_SD")
	condition = utils.StringWhereConvert(condition, search.JobEd.NullString, "t1.JOB_ED")
	condition = utils.StringWhereConvert(condition, search.JobPmNm.NullString, "t2.USER_NAME")

	var columns []string
	columns = append(columns, "t1.JNO")
	columns = append(columns, "t1.JOB_NO")
	columns = append(columns, "t1.JOB_NAME")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	var order string
	if page.Order.Valid {
		order = page.Order.String
	} else {
		order = `JNO DESC,
				CASE 
					WHEN t1.REG_DATE IS NULL THEN t1.MOD_DATE 
					WHEN t1.MOD_DATE IS NULL THEN t1.REG_DATE 
					ELSE GREATEST(t1.REG_DATE, t1.MOD_DATE) 
				END DESC NULLS LAST`
	}

	query := fmt.Sprintf(`
								SELECT *
								FROM (
									SELECT ROWNUM AS RNUM, sorted_data.*
									FROM (
									    SELECT 
											t1.JNO,
											t1.JOB_NO,
											t1.JOB_NAME,
											t1.JOB_YEAR,
											t1.JOB_SD,
											t1.JOB_ED,
											t1.COMP_NAME,
											t1.ORDER_COMP_NAME,
											t2.USER_NAME AS JOB_PM_NM,
											t2.DUTY_NAME,
											t5.CD_NM
										FROM s_job_info t1
										LEFT JOIN COMMON.V_BIZ_USER_INFO t2 ON t1.JOB_PM = t2.UNO
										LEFT JOIN TIMESHEET.SYS_CODE_SET t5 ON t5.MINOR_CD = t1.job_state AND t5.major_cd = 'JOB_STATE'
										WHERE t1.JOB_STATE = 'Y'
										AND t1.JNO NOT IN(
											SELECT JNO
											FROM IRIS_SITE_JOB
										) AND REGEXP_LIKE(job_no, '^[^-]+-[^ -]*[%s][^ -]*-[^-]+-[^-]+$')
										%s %s
										ORDER BY %s 
									) sorted_data
									WHERE ROWNUM <= :1
								)
								WHERE RNUM > :2`, typeString, condition, retryCondition, order)

	if err := db.SelectContext(ctx, &nonProjects, query, page.EndNum, page.StartNum); err != nil {
		return &nonProjects, utils.CustomErrorf(err)
	}

	return &nonProjects, nil
}

// func: 현장근태에 등록되지 않은 프로젝트 수(타입별)
// @param
// -
func (r *Repository) GetNonUsedProjectCountByType(ctx context.Context, db Queryer, search entity.NonUsedProject, retry string, typeString string) (int, error) {
	var count int

	condition := ""

	condition = utils.Int64WhereConvert(condition, search.Jno.NullInt64, "t1.JNO")
	condition = utils.StringWhereConvert(condition, search.JobNo.NullString, "t1.JOB_NO")
	condition = utils.StringWhereConvert(condition, search.JobName.NullString, "t1.JOB_NAME")
	condition = utils.Int64WhereConvert(condition, search.JobYear.NullInt64, "t1.JOB_YEAR")
	condition = utils.StringWhereConvert(condition, search.JobSd.NullString, "t1.JOB_SD")
	condition = utils.StringWhereConvert(condition, search.JobEd.NullString, "t1.JOB_ED")
	condition = utils.StringWhereConvert(condition, search.JobPmNm.NullString, "t2.USER_NAME")

	var columns []string
	columns = append(columns, "t1.JNO")
	columns = append(columns, "t1.JOB_NO")
	columns = append(columns, "t1.JOB_NAME")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	query := fmt.Sprintf(`
								SELECT 
									COUNT(*)
								FROM s_job_info t1
								LEFT JOIN COMMON.V_BIZ_USER_INFO t2 ON t1.JOB_PM = t2.UNO
								WHERE t1.JOB_STATE = 'Y'
								AND t1.JNO NOT IN(
									SELECT JNO
									FROM IRIS_SITE_JOB
								) AND REGEXP_LIKE(job_no, '^[^-]+-[^ -]*[%s][^ -]*-[^-]+-[^-]+$')
								%s %s`, typeString, condition, retryCondition)

	if err := db.GetContext(ctx, &count, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, utils.CustomErrorf(err)
	}

	return count, nil
}

// 현장별 프로젝트 조회
func (r *Repository) GetProjectBySite(ctx context.Context, db Queryer, sno int64) (entity.ProjectInfos, error) {
	var list entity.ProjectInfos

	query := `
		SELECT 
			T1.SNO,
			T1.JNO,
			T2.JOB_NAME	AS PROJECT_NM
		FROM IRIS_SITE_JOB T1
		LEFT JOIN SYS_JOB_INFO T2 ON T1.JNO = T2.JNO
		WHERE T1.SNO = :1
		ORDER BY JNO DESC`

	if err := db.SelectContext(ctx, &list, query, sno); err != nil {
		return nil, utils.CustomErrorf(err)
	}
	return list, nil
}

// func: 현장 프로젝트 추가
// @param
// -
func (r *Repository) AddProject(ctx context.Context, tx Execer, project entity.ReqProject) error {
	agent := utils.GetAgent()

	query := `
			INSERT INTO IRIS_SITE_JOB(
				SNO, JNO, IS_USE, IS_DEFAULT, REG_DATE,
				REG_AGENT, REG_USER, REG_UNO
			) VALUES (
				:1, :2, 'Y', 'N', SYSDATE,
				:3, :4, :5
			)`

	if _, err := tx.ExecContext(ctx, query, project.Sno, project.Jno, agent, project.RegUser, project.RegUno); err != nil {
		return utils.CustomErrorf(err)
	}
	return nil
}

// func: 현장 기본 프로젝트 변경
// @param
// -
func (r *Repository) ModifyDefaultProject(ctx context.Context, tx Execer, project entity.ReqProject) error {
	agent := utils.GetAgent()

	query := `
		UPDATE IRIS_SITE_JOB
		SET IS_DEFAULT = 'N',
		    MOD_AGENT = :1,
		    MOD_USER = :2,
		    MOD_UNO = :3,
		    MOD_DATE = SYSDATE
		WHERE SNO = :4`
	if _, err := tx.ExecContext(ctx, query, agent, project.ModUser, project.ModUno, project.Sno); err != nil {
		return utils.CustomErrorf(err)
	}

	query = `
		UPDATE IRIS_SITE_JOB
		SET IS_DEFAULT = 'Y',
		    MOD_AGENT = :1,
		    MOD_USER = :2,
		    MOD_UNO = :3,
		    MOD_DATE = SYSDATE
		WHERE SNO = :4
		AND JNO = :5`
	if _, err := tx.ExecContext(ctx, query, agent, project.ModUser, project.ModUno, project.Sno, project.Jno); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}

// func: 현장 프로젝트 사용여부 변경
// @param
// -
func (r *Repository) ModifyUseProject(ctx context.Context, tx Execer, project entity.ReqProject) error {
	agent := utils.GetAgent()

	query := `
		UPDATE IRIS_SITE_JOB
		SET IS_USE = :1,
		    MOD_AGENT = :2,
		    MOD_USER = :3,
		    MOD_UNO = :4,
		    MOD_DATE = SYSDATE
		WHERE SNO = :5
		AND JNO = :6`
	if _, err := tx.ExecContext(ctx, query, project.IsUsed, agent, project.ModUser, project.ModUno, project.Sno, project.Jno); err != nil {
		return utils.CustomErrorf(err)
	}
	return nil
}

// func: 현장 프로젝트 삭제
// @param
// -
func (r *Repository) RemoveProject(ctx context.Context, tx Execer, sno int64, jno int64) error {
	query := `
		DELETE FROM IRIS_SITE_JOB		
		WHERE SNO = :1
		AND JNO = :2`
	if _, err := tx.ExecContext(ctx, query, sno, jno); err != nil {
		return utils.CustomErrorf(err)
	}
	return nil
}

// func: 현장 프로젝트 사용안함 변경
// @param
// -
func (r *Repository) ModifyProjectIsNonUse(ctx context.Context, tx Execer, site entity.ReqSite) error {
	agent := utils.GetAgent()

	var jnoCondition string
	if site.Jno.Valid {
		jnoCondition = fmt.Sprintf(`AND JNO = %d`, site.Jno.Int64)
	}

	query := fmt.Sprintf(`
			UPDATE IRIS_SITE_JOB
			SET IS_USE = 'N',
			MOD_AGENT = :1,
		    MOD_USER = :2,
		    MOD_UNO = :3,
		    MOD_DATE = SYSDATE
			WHERE SNO = :4
			%s`, jnoCondition)
	if _, err := tx.ExecContext(ctx, query, agent, site.ModUser, site.ModUno, site.Sno); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}

// func: 현장 프로젝트 사용으로 변경
// @param
// -
func (r *Repository) ModifyProjectIsUse(ctx context.Context, tx Execer, site entity.ReqSite) error {
	agent := utils.GetAgent()

	var jnoCondition string
	if site.Jno.Valid {
		jnoCondition = fmt.Sprintf(`AND JNO = %d`, site.Jno.Int64)
	}

	query := fmt.Sprintf(`
			UPDATE IRIS_SITE_JOB
			SET IS_USE = 'Y',
			MOD_AGENT = :1,
		    MOD_USER = :2,
		    MOD_UNO = :3,
		    MOD_DATE = SYSDATE
			WHERE SNO = :4
			%s`, jnoCondition)
	if _, err := tx.ExecContext(ctx, query, agent, site.ModUser, site.ModUno, site.Sno); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}

// func: 현장 프로젝트 수정
// @param
// -
func (r *Repository) ModifyProject(ctx context.Context, tx Execer, project entity.ReqProject) error {
	agent := utils.GetAgent()
	query := `
			UPDATE IRIS_SITE_JOB
			SET 
				MOD_AGENT = :1,
				MOD_USER = :2,
				MOD_UNO = :3,
				MOD_DATE = SYSDATE,
				WORK_RATE = :4
			WHERE JNO = :5`
	if _, err := tx.ExecContext(ctx, query, agent, project.ModUser, project.ModUno, project.WorkRate, project.Jno); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}
