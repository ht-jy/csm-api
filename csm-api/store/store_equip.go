package store

import (
	"context"
	"csm-api/entity"
	"csm-api/utils"
	"time"
)

func (r *Repository) GetEquipList(ctx context.Context, db Queryer, jno int64, sno int64) (entity.Equips, error) {
	list := entity.Equips{}

	query := `
			SELECT 
			    T1.SNO, 
			    T1.JNO, 
			    NVL(T1.CNT, 0) AS CNT,
			    T2.JOB_NAME,
			    T1.RECORD_DATE
			FROM IRIS_EQUIP_SET T1 
			LEFT JOIN S_JOB_INFO T2 ON T1.JNO = T2.JNO
			WHERE 
			    T1.JNO = :1
				AND T1.SNO = :2 `

	if err := db.SelectContext(ctx, &list, query, jno, sno); err != nil {
		return list, utils.CustomErrorf(err)
	}
	return list, nil
}

func (r *Repository) GetEquip(ctx context.Context, db Queryer, jno int64, sno int64, recordDate time.Time) (entity.Equips, error) {
	equips := entity.Equips{}

	query := `
			SELECT 
			    e.SNO,
			    e.JNO,
			    e.CNT,
			    e.RECORD_DATE
			FROM IRIS_EQUIP_SET e
			WHERE 
				e.JNO = :1
				AND e.SNO = :2
				AND TRUNC(e.RECORD_DATE) = TRUNC(:3)
		`

	if err := db.SelectContext(ctx, &equips, query, jno, sno, recordDate); err != nil {
		return equips, utils.CustomErrorf(err)
	}

	return equips, nil

}

func (r *Repository) MergeEquipCnt(ctx context.Context, tx Execer, equip entity.Equip) error {
	agent := utils.GetAgent()

	query := `
			MERGE INTO IRIS_EQUIP_SET T1
			USING(
				SELECT
					:1 AS SNO,
					:2 AS JNO,
					:3 AS CNT,
					:4 AS RECORD_DATE,
					:5 AS USER_NAME,
					:6 AS UNO,
					:7 AS AGENT
				FROM DUAL
			) T2 
			ON (
				T1.SNO = T2.SNO
				AND T1.JNO = T2.JNO
				AND TRUNC(T1.RECORD_DATE) = TRUNC(T2.RECORD_DATE)
			) WHEN MATCHED THEN
				UPDATE SET
					T1.CNT = T2.CNT,
					T1.MOD_USER = T2.USER_NAME,
					T1.MOD_UNO = T2.UNO, 
					T1.MOD_DATE = SYSDATE,
					T1.MOD_AGENT = T2.AGENT
			WHEN NOT MATCHED THEN
				INSERT (SNO, JNO, CNT, RECORD_DATE, REG_USER, REG_UNO, REG_AGENT, REG_DATE)
				VALUES (T2.SNO, T2.JNO, T2.CNT, TRUNC(T2.RECORD_DATE), T2.USER_NAME, T2.UNO, T2.AGENT, SYSDATE)`

	if _, err := tx.ExecContext(ctx, query, equip.Sno, equip.Jno, equip.Cnt, equip.RecordDate, equip.RegUser, equip.RegUno, agent); err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}
