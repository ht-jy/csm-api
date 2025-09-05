package store

import (
	"context"
	"csm-api/entity"
	"csm-api/utils"
	"fmt"
)

// func: 휴무일 조회
// @param
// -
func (r *Repository) GetRestScheduleList(ctx context.Context, db Queryer, isRole bool, jno int64, uno string, year string, month string) (entity.RestSchedules, error) {
	list := entity.RestSchedules{}

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
		condition = fmt.Sprintf("R.JNO = %d", jno)
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
			R.CNO,
			R.JNO,
			R.IS_EVERY_YEAR,
			R.REST_YEAR,
			R.REST_MONTH,
			R.REST_DAY,
			R.REASON
		FROM IRIS_SCH_REST_SET R, USER_IN_JNO J 
		WHERE 
			R.JNO = J.JNO
			AND
			  (
				(R.IS_EVERY_YEAR = 'Y' AND TO_CHAR(R.REST_MONTH) = :1 OR :2 IS NULL)
				OR
				(R.IS_EVERY_YEAR = 'N' AND TO_CHAR(R.REST_YEAR) = :3 AND TO_CHAR(R.REST_MONTH) = :4 OR :5 IS NULL)
			  )
			AND %s
		ORDER BY JNO ASC
	`, roleCondition, condition)

	if err := db.SelectContext(ctx, &list, query, month, month, year, month, month); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return list, nil
}

// func: 휴무일 추가
// @param
// -
func (r *Repository) AddRestSchedule(ctx context.Context, tx Execer, schedule entity.RestSchedules) error {
	agent := utils.GetAgent()

	query := `
			INSERT INTO IRIS_SCH_REST_SET(
				JNO, IS_EVERY_YEAR, REST_YEAR, REST_MONTH, REST_DAY, 
			    REASON, REG_DATE, REG_AGENT, REG_UNO, REG_USER
			)  SELECT
				:1, :2, :3, :4, :5, 
				:6, SYSDATE, :8, :9, :10
				FROM dual
				WHERE NOT EXISTS(
				    SELECT 1
				    FROM IRIS_SCH_REST_SET
				    WHERE 
				    	JNO = :11 
						AND REST_YEAR = :12
						AND REST_MONTH = :13
						AND REST_DAY = :14
						AND TRIM(REASON) = TRIM(:15)
				)
			`

	for _, rest := range schedule {
		if result, err := tx.ExecContext(ctx, query,
			rest.Jno, rest.IsEveryYear, rest.RestYear, rest.RestMonth, rest.RestDay,
			rest.Reason /*SYSDATE*/, agent, rest.RegUno, rest.RegUser,
			rest.Jno, rest.RestYear, rest.RestMonth, rest.RestDay, rest.Reason,
		); err != nil {
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

// func: 휴무일 수정
// @param
// -
func (r *Repository) ModifyRestSchedule(ctx context.Context, tx Execer, schedule entity.RestSchedule) error {
	agent := utils.GetAgent()

	query := `
			UPDATE IRIS_SCH_REST_SET
			SET	
				JNO = :1,
				IS_EVERY_YEAR = :2,
				REST_YEAR = :3,
				REST_MONTH = :4,
				REST_DAY = :5,
				REASON = :6,
				MOD_DATE = SYSDATE,
				MOD_AGENT = :7,
				MOD_UNO = :8,
				MOD_USER = :9
			WHERE CNO = :10 AND
				NOT EXISTS(
						SELECT 1
						FROM IRIS_SCH_REST_SET
						WHERE 
							JNO = :11 
							AND REST_YEAR = :12
							AND REST_MONTH = :13
							AND REST_DAY = :14
							AND CNO != :15
							AND TRIM(REASON) = TRIM(:16)
					)
			`

	if result, err := tx.ExecContext(ctx, query,
		schedule.Jno, schedule.IsEveryYear, schedule.RestYear, schedule.RestMonth, schedule.RestDay, schedule.Reason, agent, schedule.ModUno, schedule.ModUser, schedule.Cno,
		schedule.Jno, schedule.RestYear, schedule.RestMonth, schedule.RestDay, schedule.Cno, schedule.Reason,
	); err != nil {
		return utils.CustomErrorf(err)
	} else {
		count, _ := result.RowsAffected()
		if count == 0 {
			return utils.CustomErrorf(fmt.Errorf("중복데이터 존재"))
		}
	}

	return nil
}

// func: 휴무일 삭제
// @param
// -
func (r *Repository) RemoveRestSchedule(ctx context.Context, tx Execer, cno int64) error {
	query := `DELETE FROM IRIS_SCH_REST_SET WHERE CNO = :1`

	if _, err := tx.ExecContext(ctx, query, cno); err != nil {
		return utils.CustomErrorf(err)
	}
	return nil
}
