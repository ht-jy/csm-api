package store

import (
	"context"
	"csm-api/entity"
	"csm-api/utils"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

/**
 * @author 작성자: 김진우
 * @created 작성일: 2025-02-12
 * @modified 최종 수정일:
 * @modifiedBy 최종 수정자:
 * @modified description
 * -
 */

// func: 현장 관리 조회
// @param
// - targetDate: 현재시간
func (r *Repository) GetSiteList(ctx context.Context, db Queryer, targetDate time.Time, role int, uno int64, status string) (*entity.Sites, error) {
	sites := entity.Sites{}

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
					t1.SITE_NM,
					t1.ETC,
					t1.LOC_CODE,
					t1.LOC_NAME,
					t1.STATUS,
					t1.REG_DATE,
					t1.REG_USER,
					t1.REG_UNO,
					t1.MOD_DATE,
					t1.MOD_USER,
					t1.MOD_UNO,
					t2.JNO AS DEFAULT_JNO,
					t3.JOB_NAME AS DEFAULT_PROJECT_NAME,
					t3.JOB_NO AS DEFAULT_PROJECT_NO,
					CASE
						WHEN (
							SELECT COUNT(*)
							FROM IRIS_WORKER_DAILY_SET d
							WHERE d.SNO = t1.SNO
							  AND TRUNC(d.RECORD_DATE) = TRUNC(:5)
							  AND d.WORK_STATE = '01'
						) >= 5 THEN 'Y'
						WHEN EXISTS (
							SELECT 1
							FROM IRIS_SCH_REST_SET r
							WHERE r.JNO = t2.JNO
							  AND TO_DATE(r.REST_YEAR || LPAD(r.REST_MONTH, 2, '0') || LPAD(r.REST_DAY, 2, '0'), 'YYYYMMDD') = TRUNC(:4)
						) THEN 'H'
						ELSE 'C'
					END AS CURRENT_SITE_STATS
				FROM IRIS_SITE_SET t1
				INNER JOIN IRIS_SITE_JOB t2 ON t1.SNO = t2.SNO AND t2.IS_DEFAULT = 'Y'
				INNER JOIN S_JOB_INFO t3 ON t2.JNO = t3.JNO AND t3.JNO IN (SELECT * FROM USER_IN_JNO)
				INNER JOIN IRIS_SITE_DATE t4 ON t1.SNO = t4.SNO AND (t1.STATUS = 'S' OR :6 >= t4.OPENING_DATE OR t4.OPENING_DATE IS NULL)
				WHERE t1.SNO > -1
				AND t1.IS_USE='Y' 
				AND t1.STATUS = :7
				ORDER BY t1.REG_DATE ASC,t1.SNO DESC`

	if err := db.SelectContext(ctx, &sites, sql, role, uno, uno, targetDate, targetDate, targetDate, status); err != nil {
		return &sites, utils.CustomErrorf(err)
	}

	return &sites, nil
}

// func: 현장 데이터 리스트
// @param
// -
func (r *Repository) GetSiteNmList(ctx context.Context, db Queryer, page entity.PageSql, search entity.Site, nonSite int) (*entity.Sites, error) {

	condition := ""

	condition = utils.Int64WhereConvert(condition, search.Sno.NullInt64, "t1.SNO")
	condition = utils.StringWhereConvert(condition, search.SiteNm.NullString, "t1.SITE_NM")
	condition = utils.StringWhereConvert(condition, search.Etc.NullString, "t1.ETC")
	condition = utils.StringWhereConvert(condition, search.LocName.NullString, "t1.LOC_NAME")

	sites := entity.Sites{}

	order := ""
	if page.Order.Valid {
		order = page.Order.String
	} else {
		order = "''"
	}
	query := fmt.Sprintf(`
				SELECT * FROM (
				    SELECT ROWNUM AS RNUM, sorted_data.*
				    FROM (				        
						SELECT 
							t1.SNO,
							t1.SITE_NM,
							t1.LOC_CODE,
							t1.LOC_NAME,
							t1.ETC,
							t1.REG_DATE,
							t1.MOD_DATE
						FROM IRIS_SITE_SET t1
						WHERE (sno > 100
							OR ( 1=:1 AND sno = 100))
							%s
				    ) sorted_data
					WHERE ROWNUM <= :2
					ORDER BY 
					    CASE WHEN 
							SNO = 100 
							THEN 0 
							ELSE 1 
						END,
					    %s,
						REG_DATE ASC, SNO DESC
				) WHERE RNUM > :3`, condition, order)

	if err := db.SelectContext(ctx, &sites, query, nonSite, page.EndNum, page.StartNum); err != nil {
		return &sites, utils.CustomErrorf(err)
	}
	return &sites, nil
}

// func: 현장 데이터 개수
// @param
// -
func (r *Repository) GetSiteNmCount(ctx context.Context, db Queryer, search entity.Site, nonSite int) (int, error) {
	var count int

	condition := ""

	condition = utils.Int64WhereConvert(condition, search.Sno.NullInt64, "t1.SNO")
	condition = utils.StringWhereConvert(condition, search.SiteNm.NullString, "t1.SITE_NM")
	condition = utils.StringWhereConvert(condition, search.Etc.NullString, "t1.ETC")
	condition = utils.StringWhereConvert(condition, search.LocName.NullString, "t1.LOC_NAME")

	query := fmt.Sprintf(`			        
						SELECT 
							count(*)
						FROM IRIS_SITE_SET t1
						WHERE (sno > 100
							OR ( 1= :1 AND sno = 100))
							%s
				    `, condition)

	if err := db.GetContext(ctx, &count, query, nonSite); err != nil {
		return 0, utils.CustomErrorf(err)
	}
	return count, nil
}

// func: 현장 상태 조회
// @param
// -
func (r *Repository) GetSiteStatsList(ctx context.Context, db Queryer, targetDate time.Time) (*entity.Sites, error) {
	sites := entity.Sites{}

	query := `
			SELECT 
				S.SNO,
				CASE
					WHEN NVL(W.WORKER_COUNT, 0) >= 5 THEN 'Y'
					WHEN H.SNO IS NOT NULL THEN 'H'
					ELSE 'C'
				END AS CURRENT_SITE_STATS
			FROM 
				(SELECT DISTINCT SNO FROM IRIS_SITE_JOB) S
				LEFT JOIN (
					SELECT SNO, COUNT(*) AS WORKER_COUNT
					FROM IRIS_WORKER_DAILY_SET
					WHERE TRUNC(RECORD_DATE) = TRUNC(:1)
					  AND WORK_STATE = '01'
					GROUP BY SNO
				) W ON S.SNO = W.SNO
				LEFT JOIN (
					SELECT DISTINCT SJ.SNO
					FROM IRIS_SITE_JOB SJ
					JOIN IRIS_SCH_REST_SET RS ON SJ.JNO = RS.JNO
					WHERE TO_DATE(RS.REST_YEAR || LPAD(RS.REST_MONTH, 2, '0') || LPAD(RS.REST_DAY, 2, '0'), 'YYYYMMDD') = TRUNC(:2)
				) H ON S.SNO = H.SNO
			ORDER BY S.SNO DESC`
	if err := db.SelectContext(ctx, &sites, query, targetDate, targetDate); err != nil {
		return &sites, utils.CustomErrorf(err)
	}
	return &sites, nil
}

// func: 현장 수정
// @param
// -
func (r *Repository) ModifySite(ctx context.Context, tx Execer, site entity.Site) error {
	agent := utils.GetAgent()

	query := `
			UPDATE IRIS_SITE_SET 
			SET
			    SITE_NM = :1,
			    ETC = :2,
				MOD_UNO = :3,
				MOD_USER = :4,
				MOD_AGENT = :5,
				MOD_DATE = SYSDATE
			WHERE
			    SNO = :6
			`
	if _, err := tx.ExecContext(ctx, query, site.SiteNm, site.Etc, site.ModUno, site.ModUser, agent, site.Sno); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}

// func: 현장 생성
// @param
// -
func (r *Repository) AddSite(ctx context.Context, db Queryer, tx Execer, jno int64, user entity.User) error {

	// IRIS_SITE_SET 생성
	query := `
			INSERT INTO IRIS_SITE_SET(
				SNO, 
			  	SITE_NM, LOC_CODE, LOC_NAME, IS_USE, STATUS,
			    REG_DATE, REG_AGENT, REG_USER, REG_UNO
			) 
			SELECT 
				SEQ_IRIS_SITE_SET.NEXTVAL,
				JOB_NAME, JOB_LOC, JOB_LOC_NAME, 'Y', 'Y'
				SYSDATE, :1, :2, :3
			FROM s_job_info 
			WHERE JNO = :4`
	if _, err := tx.ExecContext(ctx, query, user.Agent, user.UserName, user.Uno, jno); err != nil {
		return utils.CustomErrorf(err)
	}

	// IRIS_SITE_JOB 생성
	query = `
			INSERT INTO IRIS_SITE_JOB(
				SNO, 
				JNO, IS_USE, STATUS, IS_DEFAULT, REG_DATE,
				REG_AGENT, REG_USER, REG_UNO
			) VALUES (
				(SELECT SNO FROM IRIS_SITE_SET S INNER JOIN (SELECT * FROM S_JOB_INFO WHERE JNO = :1) J ON S.SITE_NM = J.JOB_NAME),
			    :2, 'Y', 'Y', 'Y', SYSDATE,
				:3, :4, :5
			)`
	if _, err := tx.ExecContext(ctx, query, jno, jno, user.Agent, user.UserName, user.Uno); err != nil {
		return utils.CustomErrorf(err)
	}

	// IRIS_SITE_DATE 생성
	query = `
			INSERT INTO IRIS_SITE_DATE(
				SNO
				, OPENING_DATE, CLOSING_PLAN_DATE, IS_USE, REG_DATE,
				REG_AGENT, REG_USER, REG_UNO
			)
			SELECT
				(SELECT SNO FROM IRIS_SITE_SET S INNER JOIN (SELECT * FROM S_JOB_INFO WHERE JNO = :1) J ON S.SITE_NM = J.JOB_NAME)
				,TO_DATE(JOB_SD, 'YYYY-MM-DD'), TO_DATE(JOB_ED, 'YYYY-MM-DD'), 'Y', SYSDATE,
				:2, :3, :4
			FROM s_job_info
			WHERE JNO = :5`
	if _, err := tx.ExecContext(ctx, query, jno, user.Agent, user.UserName, user.Uno, jno); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}

// func: 현장 사용안함 변경
// @param
// -
func (r *Repository) ModifySiteIsNonUse(ctx context.Context, tx Execer, site entity.ReqSite) error {
	agent := utils.GetAgent()

	query := `
			UPDATE IRIS_SITE_SET
			SET 
			    STATUS = 'S',
				MOD_AGENT = :1,
				MOD_USER = :2,
				MOD_UNO = :3,
				MOD_DATE = SYSDATE
			WHERE SNO = :4`
	if _, err := tx.ExecContext(ctx, query, agent, site.ModUser, site.ModUno, site.Sno); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}

// func: 현장 사용으로 변경
// @param
// -
func (r *Repository) ModifySiteIsUse(ctx context.Context, tx Execer, site entity.ReqSite) error {
	agent := utils.GetAgent()

	query := `
			UPDATE IRIS_SITE_SET
			SET 
			    STATUS = 'Y',
				MOD_AGENT = :1,
				MOD_USER = :2,
				MOD_UNO = :3,
				MOD_DATE = SYSDATE
			WHERE SNO = :4`
	if _, err := tx.ExecContext(ctx, query, agent, site.ModUser, site.ModUno, site.Sno); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}

// func: 공정률 전날 수치로 세팅
func (r *Repository) SettingWorkRate(ctx context.Context, tx Execer, targetDate time.Time) (int64, error) {
	query := `
		INSERT INTO IRIS_JOB_WORK_RATE (
			SNO, JNO, RECORD_DATE, WORK_RATE, MOD_DATE, MOD_USER, MOD_UNO
		)
		SELECT 
			T1.SNO,
			T1.JNO,
			TRUNC(:1),
			NVL(T2.WORK_RATE, 0),
			SYSDATE,
			'SYSTEM',
			0
		FROM IRIS_SITE_JOB T1
		LEFT JOIN (
			SELECT T2.SNO, T2.JNO, T2.WORK_RATE
			FROM IRIS_JOB_WORK_RATE T2
			WHERE (T2.SNO, T2.JNO, T2.RECORD_DATE) IN (
				SELECT R2.SNO, R2.JNO, MAX(R2.RECORD_DATE)
				FROM IRIS_JOB_WORK_RATE R2
				WHERE TRUNC(RECORD_DATE) < TRUNC(:2)
				GROUP BY R2.SNO, R2.JNO
			)
		) T2 ON T2.SNO = T1.SNO AND T2.JNO = T1.JNO
		WHERE NOT EXISTS (
			SELECT 1 
			FROM IRIS_JOB_WORK_RATE T3
			WHERE T3.JNO = T1.JNO
			AND TRUNC(T3.RECORD_DATE) = TRUNC(:3)
		)
		AND T1.IS_USE = 'Y'`
	result, err := tx.ExecContext(ctx, query, targetDate, targetDate, targetDate)

	if err != nil {
		return 0, utils.CustomErrorf(err)
	}
	count, _ := result.RowsAffected()

	return count, nil

}

// 공정률 수정
func (r *Repository) ModifyWorkRate(ctx context.Context, tx Execer, workRate entity.SiteWorkRate) error {
	agent := utils.GetAgent()

	query :=
		` 
			UPDATE IRIS_JOB_WORK_RATE 
			SET 
				WORK_RATE = :1,
				MOD_DATE = SYSDATE,
				MOD_UNO = :2,
				MOD_USER = :3,
				MOD_AGENT = :4,
				SNO = :5 
			WHERE SNO = :6
			AND JNO = :7
			AND TO_CHAR(RECORD_DATE, 'YYYY-MM-DD') = :8
			`
	if _, err := tx.ExecContext(ctx, query, workRate.WorkRate, workRate.ModUno, workRate.ModUser, agent, workRate.Sno, workRate.Sno, workRate.Jno, workRate.SearchDate); err != nil {
		return utils.CustomErrorf(err)
	}
	return nil
}

// 날짜별 공정률 조회
func (r *Repository) GetSiteWorkRateByDate(ctx context.Context, db Queryer, jno int64, searchDate string) (entity.SiteWorkRate, error) {
	workRate := entity.SiteWorkRate{
		WorkRate:   utils.ParseNullInt("0"),
		IsWorkRate: utils.ParseNullString("N"),
	}

	query := `
		SELECT 
			WORK_RATE, IS_WORK_RATE
		FROM (
			SELECT WORK_RATE, 'Y' AS IS_WORK_RATE
			FROM IRIS_JOB_WORK_RATE
			WHERE JNO = :1
			AND RECORD_DATE = TO_DATE(:2, 'YYYY-MM-DD')
			UNION ALL
			SELECT WORK_RATE, 'N' AS IS_WORK_RATE
			FROM IRIS_JOB_WORK_RATE
			WHERE JNO = :3
			AND RECORD_DATE = (
				SELECT MAX(RECORD_DATE)
				FROM IRIS_JOB_WORK_RATE
				WHERE JNO = :4
				AND RECORD_DATE < TO_DATE(:5, 'YYYY-MM-DD')
			)
			UNION ALL
			SELECT 0 AS WORK_RATE, 'N' AS IS_WORK_RATE FROM DUAL
		)
		WHERE ROWNUM = 1
`

	if err := db.GetContext(ctx, &workRate, query, jno, searchDate, jno, jno, searchDate); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return workRate, nil
		}
		return workRate, err
	}
	return workRate, nil
}

// 월별 공정률 조회
func (r *Repository) GetSiteWorkRateListByMonth(ctx context.Context, db Queryer, jno int64, searchDate string) (entity.SiteWorkRates, error) {
	workRates := entity.SiteWorkRates{}

	query := `
			WITH DATE_LIST AS (
			  SELECT TO_DATE(:1, 'YYYY-MM') + LEVEL - 1 AS RECORD_DATE, :2 AS JNO
			  FROM dual
			  CONNECT BY LEVEL <= LAST_DAY(TO_DATE(:3, 'YYYY-MM')) - TO_DATE(:4, 'YYYY-MM') + 1
			),
			BASE_DATA AS(
				SELECT  D.RECORD_DATE, D.JNO, R.SNO, R.WORK_RATE
				FROM DATE_LIST D LEFT JOIN IRIS_JOB_WORK_RATE R ON D.JNO = R.JNO AND D.RECORD_DATE = R.RECORD_DATE
			), 
			LATEST_DATA AS(
				SELECT * FROM 
				(	SELECT R.JNO, R.SNO, R.RECORD_DATE, NVL(R.WORK_RATE, 0) AS WORK_RATE, B.RECORD_DATE AS TARGET_DATE, ROW_NUMBER() OVER (PARTITION BY B.RECORD_DATE ORDER BY R.R.RECORD_DATE DESC) AS RN
					FROM BASE_DATA B
					LEFT JOIN IRIS_JOB_WORK_RATE R ON R.JNO = B.JNO AND R.RECORD_DATE < B.RECORD_DATE
				)
				WHERE RN = 1
			) 
			SELECT 
				B.RECORD_DATE,
				L.SNO,
				B.JNO,
				COALESCE(B.WORK_RATE, L.WORK_RATE) AS WORK_RATE,
				CASE WHEN B.WORK_RATE IS NULL THEN 'N' ELSE 'Y' END AS IS_WORK_RATE		
			FROM BASE_DATA B 
			INNER JOIN LATEST_DATA L ON B.RECORD_DATE = L.TARGET_DATE
			WHERE B.RECORD_DATE < SYSDATE
	`
	if err := db.SelectContext(ctx, &workRates, query, searchDate, jno, searchDate, searchDate); err != nil {
		return workRates, utils.CustomErrorf(err)
	}
	return workRates, nil

}

// 공정률 추가
func (r *Repository) AddWorkRate(ctx context.Context, tx Execer, workRate entity.SiteWorkRate) error {
	agent := utils.GetAgent()

	query := `
			INSERT INTO IRIS_JOB_WORK_RATE (WORK_RATE, SNO, JNO, RECORD_DATE, MOD_DATE, MOD_UNO, MOD_USER, MOD_AGENT )
			VALUES
				(:1, :2, :3, TO_DATE(:4, 'YYYY-MM-DD'), SYSDATE, :5, :6, :7)
			`

	if _, err := tx.ExecContext(ctx, query, workRate.WorkRate, workRate.Sno, workRate.Jno, workRate.SearchDate, workRate.ModUno, workRate.ModUser, agent); err != nil {
		return utils.CustomErrorf(err)
	}
	return nil
}
