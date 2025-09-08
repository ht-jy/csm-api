package store

import (
	"context"
	"csm-api/entity"
	"csm-api/utils"
)

// 관리자 로그인. 패스워드가 기술연구소 인 경우 해당 유저로 로그인
func (r *Repository) GetUserInfo(ctx context.Context, db Queryer, userId string) (entity.User, error) {
	user := entity.User{}

	sql := `
		SELECT
			T1.UNO,
			T1.USER_ID,
			T1.USER_NAME,
			T1.DEPT_NAME,
			T1.TEAM_NAME,
			CASE 
				WHEN T1.JOBDUTY_SORT_NO <= 'C030' THEN 'EXECUTIVE'
				WHEN T2.ROLE_CODE IS NOT NULL THEN T2.ROLE_CODE
				ELSE 'USER'
			END AS ROLE_CODE
		FROM
			COMMON.V_BIZ_USER_INFO T1
			LEFT JOIN IRIS_USER_ROLE_MAP T2 ON T1.UNO = T2.USER_UNO
		WHERE T1.IS_USE = 'Y'
		AND T1.USER_ID = :1`

	if err := db.GetContext(ctx, &user, sql, userId); err != nil {
		return user, utils.CustomErrorf(err)
	}
	return user, nil
}

// 직원 로그인
func (r *Repository) GetUserValid(ctx context.Context, db Queryer, userId string, userPwd string) (entity.User, error) {
	user := entity.User{}

	sql := `
		SELECT
			T1.UNO,
			T1.USER_ID,
			T1.USER_NAME,
			T1.DEPT_NAME,
			T1.TEAM_NAME,
			CASE 
				WHEN T1.JOBDUTY_SORT_NO <= 'C030' THEN 'EXECUTIVE'
				WHEN T2.ROLE_CODE IS NOT NULL THEN T2.ROLE_CODE
				ELSE 'USER'
			END AS ROLE_CODE
		FROM
			COMMON.V_BIZ_USER_INFO T1
			LEFT JOIN IRIS_USER_ROLE_MAP T2 ON T1.UNO = T2.USER_UNO
		WHERE T1.IS_USE = 'Y'
		AND T1.USER_ID = :1
		AND t1.USER_PWD = :2`

	if err := db.GetContext(ctx, &user, sql, userId, userPwd); err != nil {
		return user, utils.CustomErrorf(err)
	}
	return user, nil
}

// 협력업체 로그인
func (r *Repository) GetCompanyUserValid(ctx context.Context, db Queryer, userId string, userPwd string) (entity.CompanyInfo, error) {
	company := entity.CompanyInfo{}

	sql := `
		SELECT 
		    S.JNO,
			S.CNO,
			S.ID
		FROM 
			JOB_SUBCON_INFO S, 
			S_SYS_USER_SET U
		WHERE S.UNO = U.UNO(+)
		AND S.IS_USE = 'Y'
		AND S.ID = :1
		AND S.PW = :2`

	if err := db.GetContext(ctx, &company, sql, userId, userPwd); err != nil {
		return company, utils.CustomErrorf(err)
	}
	return company, nil
}

func (r *Repository) GetCompanyUser(ctx context.Context, db Queryer, userId string) (entity.CompanyInfo, error) {
	company := entity.CompanyInfo{}

	sql := `
		SELECT 
		    S.JNO,
			S.CNO,
			S.ID
		FROM 
			JOB_SUBCON_INFO S, 
			S_SYS_USER_SET U
		WHERE S.UNO = U.UNO(+)
		AND S.IS_USE = 'Y'
		AND S.ID = :1`

	if err := db.GetContext(ctx, &company, sql, userId); err != nil {
		return company, utils.CustomErrorf(err)
	}
	return company, nil
}
