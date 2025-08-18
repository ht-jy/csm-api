package store

import (
	"context"
	"csm-api/entity"
	"csm-api/utils"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

/**
 * @author 작성자: 김진우
 * @created 작성일: 2025-02-12
 * @modified 최종 수정일: 2025-02-21
 * @modifiedBy 최종 수정자: 정지영
 * @modified description
 * - 검색 및 정렬 조건 추가하여 데이터 조회하도록 변경
 */

// func: 근태인식기 전체 조회
// @param
// - page entity.PageSql: 현재페이지 번호, 리스트 목록 개수
func (r *Repository) GetDeviceList(ctx context.Context, db Queryer, page entity.PageSql, search entity.Device, retry string) (*entity.Devices, error) {
	list := entity.Devices{}

	condition := "AND 1=1"
	if search.DeviceNm.Valid {
		trimDeviceNm := strings.TrimSpace(search.DeviceNm.String)

		if trimDeviceNm != "" {
			condition += fmt.Sprintf(` AND LOWER(t1.DEVICE_NM) LIKE LOWER('%%%s%%')`, trimDeviceNm)
		}
	}
	if search.DeviceSn.Valid {
		trimDeviceSn := strings.TrimSpace(search.DeviceSn.String)

		if trimDeviceSn != "" {
			condition += fmt.Sprintf(` AND LOWER(t1.DEVICE_SN) LIKE LOWER('%%%s%%')`, trimDeviceSn)
		}
	}
	if search.SiteNm.Valid {
		trimSiteNm := strings.TrimSpace(search.SiteNm.String)

		if trimSiteNm != "" {
			condition += fmt.Sprintf(` AND LOWER(t2.SITE_NM) LIKE LOWER('%%%s%%')`, trimSiteNm)
		}
	}
	if search.Etc.Valid {
		trimEtc := strings.TrimSpace(search.Etc.String)

		if trimEtc != "" {
			condition += fmt.Sprintf(` AND LOWER(t1.ETC) LIKE LOWER('%%%s%%')`, trimEtc)
		}
	}
	if search.IsUse.Valid {
		trimIsUse := strings.TrimSpace(search.IsUse.String)

		if trimIsUse != "" {
			condition += fmt.Sprintf(` AND t1.IS_USE = UPPER('%s')`, trimIsUse)
		}
	}
	var columns []string
	columns = append(columns, "t2.SITE_NM")
	columns = append(columns, "t1.DEVICE_SN")
	columns = append(columns, "t1.DEVICE_NM")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	var order string
	if page.Order.Valid {
		order = page.Order.String
	} else {
		order = "NULL"
	}

	query := fmt.Sprintf(`SELECT *
				FROM (
					SELECT ROWNUM AS RNUM, sorted_data.*
					FROM (
						SELECT 
							t1.DNO,
							t1.SNO,
							t1.JNO,
							t3.JOB_NAME AS JOB_NAME,
							t2.SITE_NM,
							t1.DEVICE_SN,
							t1.DEVICE_NM,
							t1.ETC,
							t1.IS_USE,
							t1.REG_DATE AS REG_DATE,
							t1.MOD_DATE AS MOD_DATE
						FROM 
							IRIS_DEVICE_SET t1
						LEFT OUTER JOIN 
							IRIS_SITE_SET t2 
						ON 
							t1.SNO = t2.SNO
						LEFT JOIN S_JOB_INFO T3 ON T3.JNO = t1.JNO
 						WHERE 
							t1.SNO >= 100
							%s %s
						ORDER BY %s
					) sorted_data
					WHERE ROWNUM <= :1
				)
				WHERE RNUM > :2`, condition, retryCondition, order)

	if err := db.SelectContext(ctx, &list, query, page.EndNum, page.StartNum); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return &list, nil
}

// func: 근태인식기 전체 개수 조회
// @param
// -
func (r *Repository) GetDeviceListCount(ctx context.Context, db Queryer, search entity.Device, retry string) (int, error) {
	var count int

	condition := "AND 1=1"
	if search.DeviceNm.Valid {
		trimDeviceNm := strings.TrimSpace(search.DeviceNm.String)

		if trimDeviceNm != "" {
			condition += fmt.Sprintf(` AND LOWER(t1.DEVICE_NM) LIKE LOWER('%%%s%%')`, trimDeviceNm)
		}
	}
	if search.DeviceSn.Valid {
		trimDeviceSn := strings.TrimSpace(search.DeviceSn.String)

		if trimDeviceSn != "" {
			condition += fmt.Sprintf(` AND LOWER(t1.DEVICE_SN) LIKE LOWER('%%%s%%')`, trimDeviceSn)
		}
	}
	if search.SiteNm.Valid {
		trimSiteNm := strings.TrimSpace(search.SiteNm.String)

		if trimSiteNm != "" {
			condition += fmt.Sprintf(` AND LOWER(t2.SITE_NM) LIKE LOWER('%%%s%%')`, trimSiteNm)
		}
	}
	if search.Etc.Valid {
		trimEtc := strings.TrimSpace(search.Etc.String)

		if trimEtc != "" {
			condition += fmt.Sprintf(` AND LOWER(t1.ETC) LIKE LOWER('%%%s%%')`, trimEtc)
		}
	}
	if search.IsUse.Valid {
		trimIsUse := strings.TrimSpace(search.IsUse.String)

		if trimIsUse != "" {
			condition += fmt.Sprintf(` AND t1.IS_USE = UPPER('%s')`, trimIsUse)
		}
	}

	var columns []string
	columns = append(columns, "t2.SITE_NM")
	columns = append(columns, "t1.DEVICE_SN")
	columns = append(columns, "t1.DEVICE_NM")
	retryCondition := utils.RetrySearchTextConvert(retry, columns)

	query := fmt.Sprintf(`
				SELECT 
					COUNT(*) 
				FROM 
					IRIS_DEVICE_SET t1
				LEFT OUTER JOIN 
					IRIS_SITE_SET t2 
				ON 
					t1.SNO = t2.SNO
				WHERE 
					t1.SNO >= 100
					%s %s`, condition, retryCondition)

	if err := db.GetContext(ctx, &count, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, utils.CustomErrorf(err)
	}
	return count, nil
}

// func: 근태인식기 추가
// @param
// - device entity.DeviceSql: SNO, DEVICE_SN, DEVICE_NM, ETC, IS_USE, REG_USER
func (r *Repository) AddDevice(ctx context.Context, tx Execer, device entity.Device) error {
	agent := utils.GetAgent()

	query := `
				INSERT INTO IRIS_DEVICE_SET(
					DNO, 
					SNO, 
					DEVICE_SN, 
					DEVICE_NM, 
					JNO,
					ETC, 
					IS_USE, 
					REG_DATE, 
					REG_AGENT,
					REG_USER
				) VALUES (
					SEQ_IRIS_DEVICE_SET.NEXTVAL,
					:1,
				    :2,
				    :3,
				    :4,
				    :5,
				    :6,
				    SYSDATE,
				    :7,
				    :8    
				)`
	if _, err := tx.ExecContext(ctx, query, device.Sno, device.DeviceSn, device.DeviceNm, device.Jno, device.Etc, device.IsUse, agent, device.RegUser); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}

// func: 근태인식기 수정
// @param
// - device entity.DeviceSql: DNO, SNO, DEVICE_SN, DEVICE_NM, ETC, IS_USE, MOD_USER
func (r *Repository) ModifyDevice(ctx context.Context, tx Execer, device entity.Device) error {
	agent := utils.GetAgent()

	query := `
				UPDATE IRIS_DEVICE_SET 
				SET 
					SNO = :1, 
					DEVICE_SN = :2, 
					DEVICE_NM = :3, 
					ETC = :4, 
					IS_USE = :5, 
					MOD_DATE = SYSDATE,
					MOD_USER = :6,
					MOD_AGENT = :7, 
					JNO = :8
				WHERE DNO = :9`

	if _, err := tx.ExecContext(ctx, query, device.Sno, device.DeviceSn, device.DeviceNm, device.Etc, device.IsUse, device.ModUser, agent, device.Jno, device.Dno); err != nil {
		return utils.CustomErrorf(err)
	}
	return nil
}

// func: 근태인식기 삭제
// @param
// - dno sql.NullInt64: 홍채인식기 고유번호
func (r *Repository) RemoveDevice(ctx context.Context, tx Execer, dno sql.NullInt64) error {
	query := `DELETE FROM IRIS_DEVICE_SET WHERE DNO = :1`

	if _, err := tx.ExecContext(ctx, query, dno); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}

// func: 근태인식기 당일 들어온 로그 확인
// @param
func (r *Repository) GetDeviceLog(ctx context.Context, db Queryer) (*entity.RecdLogOrigins, error) {
	recodes := entity.RecdLogOrigins{}

	query := `
		SELECT IRIS_DATA FROM IRIS_RECD_LOG WHERE to_date(REG_DATE) = TRUNC(SYSDATE) `

	if err := db.SelectContext(ctx, &recodes, query); err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return &recodes, nil
}

// func: 근태인식기 미등록장치 확인
// @param
func (r *Repository) GetCheckRegistered(ctx context.Context, db Queryer, deviceName string) (int, error) {
	var count int

	query := `
			SELECT 
				COUNT(*)
			FROM
			    IRIS_DEVICE_SET
			WHERE
			    IS_USE = 'Y' 
				AND DEVICE_NM = :1
			`

	if err := db.GetContext(ctx, &count, query, deviceName); err != nil {
		return -1, utils.CustomErrorf(err)
	}

	return count, nil

}
