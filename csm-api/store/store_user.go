package store

import (
	"context"
	"csm-api/entity"
	"fmt"
	"strings"
)

func (r *Repository) GetUserInfoPmPeList(ctx context.Context, db Queryer, unoList []int) (*entity.UserPmPeInfoSqls, error) {
	userPmPeInfoSqls := entity.UserPmPeInfoSqls{}

	if len(unoList) == 0 {
		return &entity.UserPmPeInfoSqls{}, nil
	}

	placeholders := make([]string, len(unoList))
	args := make([]interface{}, len(unoList))

	for i, uno := range unoList {
		placeholder := fmt.Sprintf(":p%d", i+1)
		placeholders[i] = placeholder
		args[i] = uno
	}

	sql := fmt.Sprintf(`SELECT
    			t1.UNO,
    			t1.USER_ID,
    			t1.USER_NAME
			FROM COMMON.V_BIZ_USER_INFO t1
			WHERE t1.UNO IN (%s)`, strings.Join(placeholders, ","))

	if err := db.SelectContext(ctx, &userPmPeInfoSqls, sql, args...); err != nil {
		return nil, fmt.Errorf("GetUserInfoPmPeList fail: %w", err)
	}

	return &userPmPeInfoSqls, nil
}
