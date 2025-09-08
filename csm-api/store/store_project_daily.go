package store

import (
	"context"
	"csm-api/entity"
	"csm-api/utils"
	"database/sql"
	"fmt"
	"time"
)

// 현장관리 당일 작업 내용 조회
func (r *Repository) GetProjectDailyContentList(ctx context.Context, db Queryer, jno int64, targetDate time.Time) (*entity.ProjectDailys, error) {
	projectDailys := entity.ProjectDailys{}

	// jno 변환: 0이면 NULL 처리, 아니면 Valid 값으로 설정
	var jnoParam sql.NullInt64
	if jno != 0 {
		jnoParam = sql.NullInt64{Valid: true, Int64: jno}
	} else {
		jnoParam = sql.NullInt64{Valid: false}
	}

	// targetDate 변환: zero 값이면 NULL 처리, 아니면 Valid 값으로 설정
	var targetDateParam sql.NullTime
	if !targetDate.IsZero() {
		targetDateParam = sql.NullTime{Valid: true, Time: targetDate}
	} else {
		targetDateParam = sql.NullTime{Valid: false}
	}

	sql := `SELECT 
				t1.JNO,
				t1.IDX,
				t1.CONTENT,
				t1.IS_USE,
				t1.CONTENT_COLOR,
				t1.TARGET_DATE,
				t1.REG_DATE,
				t1.MOD_DATE,
				t1.REG_UNO,
				t1.REG_USER,
				t1.MOD_UNO,
				t1.MOD_USER
			FROM
				IRIS_DAILY_JOB t1
			WHERE
-- 				t1.IS_USE = 'Y'
-- 				AND 
			    t1.JNO = :2
				AND TO_CHAR(t1.TARGET_DATE, 'YYYY-MM-DD') = TO_CHAR(:2 , 'YYYY-MM-DD')
			ORDER BY
				NVL(t1.REG_DATE, t1.MOD_DATE) DESC`

	if err := db.SelectContext(ctx, &projectDailys, sql, jnoParam, targetDateParam); err != nil {
		return nil, utils.CustomErrorf(err)
	}
	return &projectDailys, nil
}

// 작업내용 조회
func (r *Repository) GetDailyJobList(ctx context.Context, db Queryer, isRole bool, jno int64, uno string, targetDate string) (entity.ProjectDailys, error) {
	projectDailys := entity.ProjectDailys{}

	// 프로젝트 전체조회 OR 본인이 속한 프로젝트
	roleCondition := ""
	if !isRole {
		// 협력업체는 ID로 검색해야하기 때문에
		roleCondition = fmt.Sprintf(`
									AND UNO = %s
								UNION
									SELECT JNO
									FROM JOB_SUBCON_INFO
									WHERE ID = %s
								`, uno, uno)
	}

	// 모든 스케줄 OR 선택한 프로젝트
	condition := ""
	if jno == 0 {
		condition = "1 = 1"
	} else {
		condition = fmt.Sprintf("D.JNO = %d", jno)
	}

	query := fmt.Sprintf(`
			WITH USER_IN_JNO AS (
						SELECT DISTINCT J.JNO
						FROM S_JOB_MEMBER_LIST M, IRIS_SITE_JOB J 
						WHERE 
							J.JNO = M.JNO(+)
							AND M.JNO IS NOT NULL
							AND J.IS_USE = 'Y'
							%s
						UNION
							SELECT 0 FROM DUAL
					)
			SELECT 
				D.IDX,
				D.JNO,
				D.CONTENT,
				D.CONTENT_COLOR,
				D.TARGET_DATE
			FROM IRIS_DAILY_JOB D, USER_IN_JNO J
			WHERE
				D.JNO = J.JNO
				AND TO_CHAR(TARGET_DATE, 'YYYY-MM') = :1
				AND %s
		`, roleCondition, condition)

	if err := db.SelectContext(ctx, &projectDailys, query, targetDate); err != nil {
		return entity.ProjectDailys{}, utils.CustomErrorf(err)
	}
	return projectDailys, nil
}

// 작업내용 추가
func (r *Repository) AddDailyJob(ctx context.Context, tx Execer, project entity.ProjectDailys) error {
	query := `
		INSERT INTO IRIS_DAILY_JOB(JNO, CONTENT, CONTENT_COLOR,TARGET_DATE, REG_DATE, REG_UNO, REG_USER)
			SELECT
				:1, :2, :3, :4, SYSDATE, :5, :6
			FROM dual
			WHERE NOT EXISTS (
				SELECT 1
				FROM IRIS_DAILY_JOB
				WHERE 
					JNO = :7
					AND TRUNC(TARGET_DATE) = TRUNC(:8)
					AND TRIM(CONTENT) = TRIM(:9)
			)
		`

	for _, job := range project {
		if result, err := tx.ExecContext(ctx, query, job.Jno, job.Content, job.ContentColor, job.TargetDate, job.RegUno, job.RegUser, job.Jno, job.TargetDate, job.Content); err != nil {
			return utils.CustomErrorf(err)
		} else {
			count, _ := result.RowsAffected()
			if count == 0 {
				return utils.CustomErrorf(fmt.Errorf("중복데이터 존재"))
			}
		}
	}

	return nil
}

// 작업내용 수정
func (r *Repository) ModifyDailyJob(ctx context.Context, tx Execer, project entity.ProjectDaily) error {
	query := `
			UPDATE IRIS_DAILY_JOB 
			SET 
				JNO = :1,
				CONTENT = :2,
				CONTENT_COLOR =:3 ,
				TARGET_DATE = :4,
				MOD_DATE = SYSDATE,
				MOD_UNO = :5,
				MOD_USER = :6
			WHERE 
			    IDX = :7
				AND NOT EXISTS (
					SELECT	1
					FROM IRIS_DAILY_JOB
					WHERE
						JNO = :8
						AND TRUNC(TARGET_DATE) = TRUNC(:9)
						AND TRIM(CONTENT) = TRIM(:10)
						AND IDX != :11			
					)
			`

	if result, err := tx.ExecContext(ctx, query, project.Jno, project.Content, project.ContentColor, project.TargetDate, project.RegUno, project.RegUser, project.Idx, project.Jno, project.TargetDate, project.Content, project.Idx); err != nil {
		return utils.CustomErrorf(err)
	} else {
		count, _ := result.RowsAffected()
		if count == 0 {
			return utils.CustomErrorf(fmt.Errorf("중복데이터 존재"))
		}
	}

	return nil
}

// 작업내용 삭제
func (r *Repository) RemoveDailyJob(ctx context.Context, tx Execer, idx int64) error {
	query := `DELETE FROM IRIS_DAILY_JOB WHERE IDX = :1`

	if _, err := tx.ExecContext(ctx, query, idx); err != nil {
		return utils.CustomErrorf(err)
	}
	return nil
}
