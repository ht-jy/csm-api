package store

import (
	"context"
	"csm-api/entity"
	"csm-api/utils"
	"fmt"
)

// func: 프로젝트에 설정된 공수 조회
// @param
// - jno: 프로젝트pk
func (r *Repository) GetManHourList(ctx context.Context, db Queryer, jno int64) (*entity.ManHours, error) {
	manHours := entity.ManHours{}

	query := `
			SELECT
			    MHNO,
			    WORK_HOUR,
			    MAN_HOUR,
			    JNO,
			    ETC
			FROM 
			    IRIS_MAN_HOUR MH
			WHERE
				MH.JNO = :1
			ORDER BY
			    WORK_HOUR DESC
		`

	if err := db.SelectContext(ctx, &manHours, query, jno); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return &manHours, nil
}

// func: 공수 수정 및 추가(시스템 관리자)
// @param
// - manHour: 공수 정보
func (r *Repository) MergeManHour(ctx context.Context, tx Execer, manHour entity.ManHour) (count int64, err error) {
	query := `
		MERGE INTO SAFE.IRIS_MAN_HOUR J1
		USING (
			SELECT 
				:1 AS MHNO,
				:2 AS WORK_HOUR,
				:3 AS MAN_HOUR,
				:4 AS JNO, 
				:5 AS ETC,
				:6 AS UNO,	
				:7 AS USER_NAME
			FROM DUAL
		) J2
		ON (
			J1.MHNO = J2.MHNO
		) WHEN MATCHED THEN
			UPDATE SET
				J1.WORK_HOUR = J2.WORK_HOUR,
				J1.MAN_HOUR = J2.MAN_HOUR,
				J1.JNO = J2.JNO,
				J1.ETC = J2.ETC,
				J1.MOD_UNO = J2.UNO,	
				J1.MOD_USER = J2.USER_NAME,
				J1.MOD_DATE = SYSDATE
		WHEN NOT MATCHED THEN
			INSERT ( WORK_HOUR, MAN_HOUR, JNO, ETC, REG_UNO, REG_USER, REG_DATE )
			VALUES (
				J2.WORK_HOUR,
				J2.MAN_HOUR,
				J2.JNO,
				J2.ETC,
				J2.UNO,	
				J2.USER_NAME,
				SYSDATE
			)
		`
	result, err := tx.ExecContext(ctx, query, manHour.Mhno, manHour.WorkHour, manHour.ManHour, manHour.Jno, manHour.Etc, manHour.RegUno, manHour.RegUser)
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}

	count, _ = result.RowsAffected()

	return count, nil
}

// func: 공수 추가
// @param
// - manHour: 공수 정보
func (r *Repository) AddManHour(ctx context.Context, tx Execer, manHour entity.ManHour) error {
	query := `
			INSERT INTO IRIS_MAN_HOUR ( WORK_HOUR, MAN_HOUR, JNO, ETC, REG_UNO, REG_USER, REG_DATE )
			VALUES (:1, :2,	:3,	:4,	:5,	:6,	SYSDATE )
		`
	_, err := tx.ExecContext(ctx, query, manHour.WorkHour, manHour.ManHour, manHour.Jno, manHour.Etc, manHour.RegUno, manHour.RegUser)
	if err != nil {
		return utils.CustomErrorf(err)
	}
	return nil
}

// func: 프로젝트 설정 정보 수정
// @param: ProjectSetting
// -
func (r *Repository) MergeProjectSetting(ctx context.Context, tx Execer, project entity.ProjectSetting) (int64, error) {
	//agent := utils.GetAgent()

	query := `
				MERGE INTO IRIS_JOB_SET J1
				USING (
					SELECT 
						:1 AS JNO,
						:2 AS IN_TIME,
						:3 AS OUT_TIME,
						:4 AS RESPITE_TIME,
						:5 AS CANCEL_CODE,
						:6 AS UNO,	
						:7 AS USER_NAME
					FROM DUAL
				) J2
				ON (
					J1.JNO = J2.JNO
				) WHEN MATCHED THEN
					UPDATE SET
						J1.IN_TIME = J2.IN_TIME,
						J1.OUT_TIME = J2.OUT_TIME,
						J1.RESPITE_TIME = J2.RESPITE_TIME,
						J1.CANCEL_CODE = J2.CANCEL_CODE,
						J1.MOD_UNO = J2.UNO,	
						J1.MOD_USER = J2.USER_NAME,
						J1.MOD_DATE = SYSDATE
				WHEN NOT MATCHED THEN
					INSERT ( JNO, IN_TIME, OUT_TIME, RESPITE_TIME, CANCEL_CODE, REG_UNO, REG_USER, REG_DATE )
					VALUES (
						J2.JNO,
						J2.IN_TIME,
						J2.OUT_TIME,
						J2.RESPITE_TIME,
						J2.CANCEL_CODE,
						J2.UNO,	
						J2.USER_NAME,
						SYSDATE		
			)`
	result, err := tx.ExecContext(ctx, query, project.Jno, project.InTime, project.OutTime, project.RespiteTime, project.CancelCode, project.RegUno, project.RegUser)
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}

	count, _ := result.RowsAffected()

	return count, nil
}

// func: 프로젝트 미설정 정보 조회(스케줄러)
// @param
// -
func (r *Repository) GetCheckProjectSetting(ctx context.Context, db Queryer) (projects *entity.ProjectSettings, err error) {
	projects = &entity.ProjectSettings{}

	query := `
				SELECT 
				    DISTINCT(JNO) 
				FROM 
				    IRIS_SITE_JOB 
				WHERE 
				    JNO NOT IN (SELECT JNO FROM IRIS_JOB_SET)
				AND IS_USE = 'Y'`

	if err = db.SelectContext(ctx, projects, query); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return
}

// func: 기본 공수 미설정 정보 조회(스케줄러)
// @param
// -
func (r *Repository) GetCheckProjectManHours(ctx context.Context, db Queryer) (projects *entity.ProjectSettings, err error) {
	projects = &entity.ProjectSettings{}

	query := `
				SELECT 
				    DISTINCT(JNO) 
				FROM 
				    IRIS_SITE_JOB 
				WHERE 
				    JNO NOT IN (SELECT JNO FROM IRIS_MAN_HOUR)`

	if err = db.SelectContext(ctx, projects, query); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return
}

// func: 프로젝트 기본 설정 정보 조회
// @param
// - jno
func (r *Repository) GetProjectSetting(ctx context.Context, db Queryer, isRole bool, uno string, jno int64) (*entity.ProjectSettings, error) {

	roleCondition := ""
	if isRole {
		roleCondition = "1=1"
	} else {
		roleCondition = fmt.Sprintf("UNO = %s", uno)
	}
	setting := entity.ProjectSettings{}
	query := fmt.Sprintf(`
			WITH USER_IN_JNO AS (
					SELECT 
						DISTINCT(JNO) 
					FROM 
						S_JOB_MEMBER_LIST M 
					WHERE
						%s
			)
			SELECT 
				J.JNO,
				J.IN_TIME,
				J.OUT_TIME,
				J.RESPITE_TIME,
				J.CANCEL_CODE,
				C.UDF_VAL_03 AS CANCEL_DAY,
				J.REG_DATE,
				J.REG_UNO,
				J.REG_USER,
				J.MOD_DATE,
				J.MOD_UNO,
				J.MOD_USER
			FROM IRIS_JOB_SET J, IRIS_CODE_SET C, USER_IN_JNO I
			WHERE
			    J.CANCEL_CODE = C.CODE
				AND J.JNO = I.JNO
				AND J.JNO = :1
			`, roleCondition)

	if err := db.SelectContext(ctx, &setting, query, jno); err != nil {
		return &setting, utils.CustomErrorf(err)
	}

	return &setting, nil
}

// func: 공수 삭제
// @param
// - mhno: 공수pk
func (r *Repository) DeleteManHour(ctx context.Context, tx Execer, mhno int64) error {
	query := fmt.Sprintf(`
			DELETE 
			FROM IRIS_MAN_HOUR
			WHERE 
			    MHNO = :1
			`)
	result, err := tx.ExecContext(ctx, query, mhno)
	if err != nil {
		return utils.CustomErrorf(err)
	} else if count, _ := result.RowsAffected(); count <= 0 {
		return utils.CustomErrorf(fmt.Errorf("Deleted ManHour is Zero"))
	}

	return nil
}

// 프로젝트 설정 저장 로그
func (r *Repository) ProjectSettingLog(ctx context.Context, tx Execer, setting entity.ProjectSetting) error {
	agent := utils.GetAgent()

	query := fmt.Sprintf(`
		INSERT INTO IRIS_JOB_MAN_HOUR_LOG( JNO, CHANGE_SETTING, MESSAGE, REG_DATE, REG_USER, REG_UNO, REG_AGENT)
		VALUES (:1, 'IRIS_JOB_SET', :2, SYSDATE, :3, :4, :5)
	`)

	if _, err := tx.ExecContext(ctx, query, setting.Jno, setting.Message, setting.RegUser, setting.RegUno, agent); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}

// 공수 설정 저장 로그
func (r *Repository) ManHourLog(ctx context.Context, tx Execer, manhour entity.ManHour) error {
	agent := utils.GetAgent()

	query := fmt.Sprintf(`
		INSERT INTO IRIS_JOB_MAN_HOUR_LOG( JNO, CHANGE_SETTING, MESSAGE, REG_DATE, REG_USER, REG_UNO, REG_AGENT)
		VALUES (:1, 'IRIS_MAN_HOUR', :2, SYSDATE, :3, :4, :5)
	`)

	if _, err := tx.ExecContext(ctx, query, manhour.Jno, manhour.Message, manhour.RegUser, manhour.RegUno, agent); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}
