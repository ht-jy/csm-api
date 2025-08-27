package store

import (
	"context"
	"csm-api/entity"
	"csm-api/utils"
	"database/sql"
	"errors"
	"fmt"
)

func (r *Repository) GetSiteDateData(ctx context.Context, db Queryer, sno int64) (*entity.SiteDate, error) {
	siteDate := entity.SiteDate{}

	query := `SELECT
				t1.OPENING_DATE,
				t1.CLOSING_PLAN_DATE,
				t1.CLOSING_FORECAST_DATE,
				t1.CLOSING_ACTUAL_DATE,
				t1.REG_UNO,
				t1.REG_USER,
				t1.REG_DATE
			FROM
				IRIS_SITE_DATE t1
			WHERE
				t1.SNO = :1
-- 				AND t1.IS_USE = 'Y'`

	if err := db.GetContext(ctx, &siteDate, query, sno); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &siteDate, nil
		}
		return nil, utils.CustomErrorf(err)
	}
	return &siteDate, nil
}

// 현장 날짜 테이블 수정
//
// @param
// - sno: 현장고유번호
// - siteDate: 현장 시간 (opening_date, closing_plan_date, closing_forecast_date, closing_actual_date)
func (r *Repository) ModifySiteDate(ctx context.Context, tx Execer, sno int64, siteDateSql entity.SiteDate) error {
	query := fmt.Sprintf(`
			UPDATE IRIS_SITE_DATE 
			SET
			    OPENING_DATE = :1,
				CLOSING_PLAN_DATE = :2,
				CLOSING_FORECAST_DATE = :3,
				CLOSING_ACTUAL_DATE = :4
			WHERE
				SNO = :5 
				AND IS_USE = 'Y'
			`)

	if _, err := tx.ExecContext(ctx, query, siteDateSql.OpeningDate, siteDateSql.ClosingPlanDate, siteDateSql.ClosingForecastDate, siteDateSql.ClosingActualDate, sno); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}

// func: 현장 날짜 사용안함 변경
// @param
// -
func (r *Repository) ModifySiteDateIsNonUse(ctx context.Context, tx Execer, site entity.ReqSite) error {
	agent := utils.GetAgent()
	query := `
			UPDATE IRIS_SITE_DATE
			SET 
			    IS_USE = 'N',
			    CLOSING_ACTUAL_DATE = (SELECT NVL(CLOSING_ACTUAL_DATE, SYSDATE) FROM IRIS_SITE_DATE WHERE SNO = :1),
				MOD_AGENT = :2,
				MOD_USER = :3,
				MOD_UNO = :4,
				MOD_DATE = SYSDATE
			WHERE SNO = :5`
	if _, err := tx.ExecContext(ctx, query, site.Sno, agent, site.ModUser, site.ModUno, site.Sno); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}

// func: 현장 날짜 사용으로 변경
// @param
// -
func (r *Repository) ModifySiteDateIsUse(ctx context.Context, tx Execer, site entity.ReqSite) error {
	agent := utils.GetAgent()
	query := `
			UPDATE IRIS_SITE_DATE
			SET 
			    IS_USE = 'Y',
			    CLOSING_ACTUAL_DATE = NULL,
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
