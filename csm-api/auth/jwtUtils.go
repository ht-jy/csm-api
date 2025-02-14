package auth

import (
	"context"
	"csm-api/clock"
	"csm-api/config"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

type JWTUtils struct {
	Cfg   *config.JwtConfig
	Clock clock.Clocker
}

type JWTRole string

const (
	Admin JWTRole = "admin"
	User  JWTRole = "user"
)

// claims 정의
type JWTClaims struct {
	UserId   string  `json:"user_id"`
	UserName string  `json:"user_name"`
	Role     JWTRole `json:"role"`
	Token    string
	IsSaved  bool
}

// context에 저장하는 claims value
type UserId struct{}
type Role struct{}

// JWTUtils 구조체 생성
func JwtNew(c clock.Clocker) (*JWTUtils, error) {
	jwt := &JWTUtils{}

	jwtConfig, err := config.GetJwtConfig()
	if err != nil {
		return nil, err
	}

	jwt.Cfg = jwtConfig
	jwt.Clock = c

	return jwt, nil
}

// 토큰 생성
func (j *JWTUtils) GenerateToken(jwtClaims *JWTClaims) (string, error) {
	// 비밀 키 선택 (아이디 저장 여부에 따라 다름)
	var secretKey []byte
	if jwtClaims.IsSaved {
		secretKey = []byte(j.Cfg.SavedSecretKey) // "아이디 저장"한 경우 다른 키 사용
	} else {
		secretKey = []byte(j.Cfg.SecretKey)
	}

	// JWT 클레임 설정
	claims := jwt.MapClaims{
		"userId":   jwtClaims.UserId,
		"userName": jwtClaims.UserName,
		"role":     jwtClaims.Role,
		"isSaved":  jwtClaims.IsSaved, // "아이디 저장" 여부 추가
	}

	// 만료 시간 설정 (아이디 저장 안 한 경우 1시간 후 만료)
	if !jwtClaims.IsSaved {
		claims["exp"] = j.Clock.Now().Add(1 * time.Hour).Unix()
	}

	// 토큰 생성
	parseToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 비밀 키로 서명
	tokenString, err := parseToken.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("jwtUtils.go/GenerateToken() err: %w", err)
	}

	jwtClaims.Token = tokenString

	return tokenString, nil
}

// 토큰 유효성 검사
func (j *JWTUtils) ValidateJWT(r *http.Request) (*JWTClaims, error) {
	// 쿠키 읽기
	cookie, err := r.Cookie("jwt")
	if err != nil {
		return nil, fmt.Errorf("jwtUtils.go/validateJWT() err: %v", err)
	}
	tokenString := cookie.Value

	// 토큰 파싱 및 검증
	parseToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 서명 방법 확인
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// "아이디 저장" 여부 확인 후 적절한 키 반환
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("invalid token claims")
		}

		if isSaved, ok := claims["isSaved"].(bool); ok && isSaved {
			return []byte(j.Cfg.SavedSecretKey), nil // "아이디 저장"한 경우 SavedSecretKey 사용
		}
		return []byte(j.Cfg.SecretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwtUtils.go/invalid token: %v", err)
	}

	// 클레임 확인
	claims, ok := parseToken.Claims.(jwt.MapClaims)
	if !ok || !parseToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// JWTClaims 매핑
	jwtClaims := &JWTClaims{
		UserId:   claims["userId"].(string),
		UserName: claims["userName"].(string),
		IsSaved:  claims["isSaved"].(bool), // "아이디 저장" 여부 확인
	}

	// 역할(Role) 처리
	roleVal, exists := claims["role"]
	if exists && roleVal != nil {
		if roleStr, ok := roleVal.(string); ok {
			switch JWTRole(roleStr) {
			case Admin, User:
				jwtClaims.Role = JWTRole(roleStr)
			}
		}
	}

	return jwtClaims, nil
}

// 사용자 토큰 인증시 필요한 api 호출시 token의 claims에 있는 값을 context에 저장
func (j *JWTUtils) FillContext(r *http.Request) (*http.Request, *JWTClaims, error) {
	// 토큰 검사 및 claims 추출
	claims, err := j.ValidateJWT(r)
	if err != nil {
		return nil, &JWTClaims{}, err
	}

	// claims 데이터 context에 저장
	ctx := SetContext(r.Context(), UserId{}, claims.UserId)
	ctx = SetContext(ctx, Role{}, string(claims.Role))

	httpRequestClone := r.Clone(ctx)

	return httpRequestClone, claims, nil
}

func SetContext(ctx context.Context, key struct{}, value string) context.Context {
	return context.WithValue(ctx, key, value)
}

func GetContext(ctx context.Context, key interface{}) (string, bool) {
	value, ok := ctx.Value(key).(string)
	return value, ok
}

// 쿠키 만료 시간 설정 (아이디 저장 여부에 따라)
func GetCookieMaxAge(isSaved bool) int {
	if isSaved {
		return 60 * 60 * 24 * 30 // "아이디 저장"하면 30일 저장
	}
	return 60 * 60 // "아이디 저장" 안 하면 1시간 후 만료
}
