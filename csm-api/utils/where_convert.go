package utils

import (
	"database/sql"
	"fmt"
	"strings"
)

func StringWhereConvert(condition string, sqlValue sql.NullString, target string) string {
	if sqlValue.Valid {
		value := strings.TrimSpace(sqlValue.String)
		if value != "" {
			condition += fmt.Sprintf(` AND LOWER(%s) LIKE LOWER('%%%s%%')`, target, value)
		}
	}
	return condition
}

func Int64WhereConvert(condition string, sqlValue sql.NullInt64, target string) string {
	if sqlValue.Valid {
		value := sqlValue.Int64
		if value != 0 {
			condition += fmt.Sprintf(` AND %s = %d`, target, value)
		}
	}
	return condition
}

func TimeWhereConvert(condition string, sqlValue sql.NullString, target string) string {
	if sqlValue.Valid {
		value := strings.TrimSpace(sqlValue.String)
		if value != "" {
			condition += fmt.Sprintf(` AND TO_CHAR(%s, 'YYYY-MM-DD') = '%s'`, target, value)
		}
	}
	return condition
}

func TimeBetweenWhereConvert(condition string, sqlValue1 sql.NullString, sqlValue2 sql.NullString, target string) string {
	if sqlValue1.Valid && sqlValue2.Valid {
		value1 := strings.TrimSpace(sqlValue1.String)
		value2 := strings.TrimSpace(sqlValue2.String)

		if value1 != "" && value2 != "" {
			condition += fmt.Sprintf(` AND TO_CHAR(%s, 'YYYY-MM-DD') BETWEEN '%s' AND '%s'`, target, value1, value2)
		}
	}
	return condition
}

func OrTimeBetweenWhereConvert(condition string, sqlValue1 sql.NullString, sqlValue2 sql.NullString, target string) string {
	if sqlValue1.Valid && sqlValue2.Valid {
		value1 := strings.TrimSpace(sqlValue1.String)
		value2 := strings.TrimSpace(sqlValue2.String)

		if value1 != "" && value2 != "" {
			condition += fmt.Sprintf(` OR TO_CHAR(%s, 'YYYY-MM-DD') BETWEEN '%s' AND '%s'`, target, value1, value2)
		}
	}
	return condition
}

// 테이블 컬럼(columns)과 검색 쿼리(retry)를 받아 SQL WHERE절(AND, OR 조건 조각)로 변환
// parameter
// - columns: ex) []string{"T1.REASON_TYPE", "T2.USER_NM", "T2.USER_ID"}
// - retry: ex) "ALL:마감?USER_NM?USER_ID;03?REASON_TYPE;07?REASON_TYPE;08?REASON_TYPE|테스트~USER_ID:010123|ALL"
// role
//  1. `~` (틸드) : 검색 컬럼을 나누는 최상위 구분자. 각 블록은 AND로 결합됨. ex) "A~B"  →  A ... AND B ...
//  2. `|` (파이프) : 하나의 블록 내에서 검색 값 또는 값의 그룹을 나누는 구분자. 각 그룹은 AND로 결합됨 (괄호로 감싸지지 않음). ex) "A|B"  →  A... AND B...
//  3. `;` (세미콜론) : 검색값 중에서 값의 그룹인 것의 "서브 조건"을 나누는 구분자. 각 서브조건은 OR로 결합됨 (괄호로 감싸짐). ex) "A;B"  →  (A... OR B...)
//  4. `:` (콜론) : "필드:값" 구문.
//     - 필드가 ALL이면 특별 규칙, 그 외에는 해당 필드명으로만 검색
//  5. `?` (물음표) : 서브조건에서 뒤에 필드명을 붙이면, 지정된 필드에서만 LIKE 검색
//     - 여러 필드명 지정 가능 ("값?USER_ID?USER_NM")
//     - 지정 없으면 모든 columns에 대해 검색
//  6. retry에 있는 검색 컬럼이 columns에 없으면 해당 필드는 where절 변환에서 무시
//
// return
// ex):
//
//	AND (LOWER(T2.USER_NM) LIKE LOWER('%마감%')
//	     OR LOWER(T2.USER_ID) LIKE LOWER('%마감%')
//	     OR LOWER(T1.REASON_TYPE) LIKE LOWER('%03%')
//	     OR LOWER(T1.REASON_TYPE) LIKE LOWER('%07%')
//	     OR LOWER(T1.REASON_TYPE) LIKE LOWER('%08%'))
//	AND (LOWER(T1.REASON_TYPE) LIKE LOWER('%테스트%')
//	     OR LOWER(T2.USER_NM) LIKE LOWER('%테스트%')
//	     OR LOWER(T2.USER_ID) LIKE LOWER('%테스트%'))
//	AND LOWER(T2.USER_ID) LIKE LOWER('%010123%')
//	AND (LOWER(T1.REASON_TYPE) LIKE LOWER('%ALL%')
//	     OR LOWER(T2.USER_NM) LIKE LOWER('%ALL%')
//	     OR LOWER(T2.USER_ID) LIKE LOWER('%ALL%'))
func RetrySearchTextConvert(retry string, columns []string) string {
	where := ""
	trimRetry := strings.TrimSpace(retry)
	if trimRetry == "" {
		return ""
	}
	andGroups := strings.Split(trimRetry, "~")
	for _, andGroup := range andGroups {
		andGroup = strings.TrimSpace(andGroup)
		if andGroup == "" {
			continue
		}
		colonIdx := strings.Index(andGroup, ":")
		if colonIdx != -1 {
			key := strings.TrimSpace(andGroup[:colonIdx])
			value := strings.TrimSpace(andGroup[colonIdx+1:])
			if strings.ToUpper(key) == "ALL" {
				// *** ALL은 원래 코드 절대 그대로! (수정 금지) ***
				orBlocks := strings.Split(value, ";")
				var orExpr []string
				for _, block := range orBlocks {
					block = strings.TrimSpace(block)
					if block == "" {
						continue
					}
					parts := strings.Split(block, "?")
					searchWord := parts[0]
					var fieldTargets []string
					isAllColumn := false
					if len(parts) > 1 {
						fieldTargets = parts[1:]
					} else {
						fieldTargets = columns
						isAllColumn = true
					}
					var temp []string
					for _, word := range strings.Split(searchWord, "|") { // 전체검색 시 |로 구분되는 값이 조회 되지 않아 코드 추가
						for _, column := range columns {
							if isAllColumn {
								temp = append(temp, fmt.Sprintf(`LOWER(%s) LIKE LOWER('%%%s%%')`, column, word))
							} else {
								for _, f := range fieldTargets {
									if strings.HasSuffix(column, f) {
										temp = append(temp, fmt.Sprintf(`LOWER(%s) LIKE LOWER('%%%s%%')`, column, word))
										break
									}
								}
							}
						}
					}
					if len(temp) > 0 {
						orExpr = append(orExpr, strings.Join(temp, " OR "))
					}
				}
				if len(orExpr) > 0 {
					where += "AND (" + strings.Join(orExpr, " OR ") + ") "
				}
			} else {
				// *** key:필드 | 여러 값은 AND로 묶기 ***
				values := strings.Split(value, "|")
				for _, v := range values {
					v = strings.TrimSpace(v)
					for _, column := range columns {
						if strings.HasSuffix(column, key) {
							where += fmt.Sprintf("AND LOWER(%s) LIKE LOWER('%%%s%%') ", column, v)
						}
					}
				}
			}
		} else {
			// *** : 없는 건 전체 columns OR ***
			var temp []string
			for _, column := range columns {
				temp = append(temp, fmt.Sprintf(`LOWER(%s) LIKE LOWER('%%%s%%')`, column, andGroup))
			}
			if len(temp) > 0 {
				where += "AND (" + strings.Join(temp, " OR ") + ") "
			}
		}
	}
	return strings.TrimSpace(where)
}
