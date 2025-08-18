package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 공백 제거 + 소문자 변환 + 특수문자 제거
func NormalizeForEqual(s string) string {
	re := regexp.MustCompile(`[^가-힣a-zA-Z0-9]`) // 한글, 영문, 숫자 외 제거
	return strings.ToLower(re.ReplaceAllString(strings.TrimSpace(s), ""))
}

// yyyy-mm-dd -> yy-mm-dd
func ConvertYYYYMMDDToYYMMDD(input string) string {
	t, _ := time.Parse("2006-01-02", strings.TrimSpace(input))
	return t.Format("06-01-02")
}

// mm-dd-yy -> yy-mm-dd
func ConvertMMDDYYToYYMMDD(input string) string {
	t, _ := time.Parse("01-02-06", strings.TrimSpace(input))
	return t.Format("06-01-02")
}

func IsYYYYMMDD(s string) bool {
	re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	return re.MatchString(s)
}

func NormalizeYYMMDD(s string) string {
	parts := strings.Split(s, "-")
	if len(parts) != 3 {
		return s // 형식 안 맞으면 그대로 반환
	}

	// 현재 연도 기준 앞 두 자리 (예: 2025 → "20")
	prefix := time.Now().Format("2006")[:2]

	fullDate := fmt.Sprintf("%s%s-%s-%s", prefix, parts[0], parts[1], parts[2])

	// 검증: 실제 날짜로 파싱 가능한지 확인
	if _, err := time.Parse("2006-01-02", fullDate); err != nil {
		return s // 파싱 실패하면 원본 그대로 반환
	}

	return fullDate
}

func NormalizeHHMM(s string) string {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return s // 형식이 아니면 그대로 반환
	}

	hour := parts[0]
	min := parts[1]

	// 시, 분을 두 자리로 맞추고 초 추가
	return fmt.Sprintf("%02s:%02s:00", hour, min)
}

func NormalizeBool(s string) bool {
	reg := regexp.MustCompile(`[^a-zA-Z]`) // 한글, 영문, 숫자 외 제거
	lower := strings.ToLower(reg.ReplaceAllString(strings.TrimSpace(s), ""))
	result, _ := strconv.ParseBool(lower)
	return result
}
