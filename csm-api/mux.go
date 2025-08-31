package main

import (
	"context"
	"csm-api/auth"
	"csm-api/clock"
	"csm-api/config"
	"csm-api/handler"
	"csm-api/route"
	"csm-api/store"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rs/cors"
	"net/http"
)

/**
 * @author 작성자: 김진우
 * @created 작성일: 2025-02-12
 * @modified 최종 수정일: 2025-02-21
 * @modifiedBy 최종 수정자: 정지영
 * @modified description
 * - 공지사항 기능 추가
 */

// func:
// @param
// - cfg *config.DBConfigs: db 접속 정보
// chi패키지를 이용하여 http method에 따른 여러 요청을 라우팅 할 수 있음 함수 구현
func newMux(ctx context.Context, safeDb *sqlx.DB, timesheetDb *sqlx.DB) (http.Handler, error) {
	mux := chi.NewRouter()
	mux.Use(handler.Recoverer)

	// CORS 미들웨어 설정
	c := cors.New(cors.Options{ // 허용할 도메인
		AllowedOrigins: []string{
			"http://localhost:3002",
			"http://127.0.0.1:3002",
			"https://61.41.17.36",
			"https://csm.htenc.co.kr",
		},
		AllowCredentials: true,                                                // 쿠키 허용
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // 허용할 메서드
		AllowedHeaders:   []string{"Content-Type", "Authorization"},           // 허용할 헤더
	})
	r := store.Repository{Clocker: clock.RealClock{}}

	// jwt struct 생성
	jwt, err := auth.JwtNew(clock.RealClock{})
	if err != nil {
		return nil, err
	}

	// api config 생성
	apiCfg, err := config.GetApiConfig()
	if err != nil {
		return nil, err
	}

	cfg, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	mux.Route("/csm", func(csm chi.Router) {
		// 공개 라우팅
		csm.Mount("/login", route.LoginRoute(jwt, safeDb, timesheetDb, &r)) // 로그인
		csm.Mount("/logout", route.LogoutRoute())                           // 로그아웃
		csm.Mount("/jwt-validation", route.JwtVaildRoute(jwt))              // jwt 유효성 검사
		//csm.Mount("/init", route.InitApiRoute(safeDb, timesheetDb, apiCfg, cfg, &r)) // api로 초기 세팅

		// 인증 라우팅
		csm.Group(func(router chi.Router) {
			router.Use(handler.AuthMiddleware(jwt))                                          // jwt 인증
			router.Mount("/menu", route.MenuRoute(safeDb, &r))                               // 메뉴
			router.Mount("/user", route.UserRoute(safeDb, timesheetDb, &r))                  // 사용자 {권한}
			router.Mount("/api", route.ApiRoute(apiCfg, safeDb, &r))                         // api
			router.Mount("/excel", route.ExcelRoute(safeDb, &r))                             // 엑셀
			router.Mount("/project", route.ProjectRoute(safeDb, timesheetDb, &r))            // 프로젝트
			router.Mount("/organization", route.OrganiztionRoute(timesheetDb, &r))           // 조직도
			router.Mount("/site", route.SiteRoute(safeDb, timesheetDb, &r, apiCfg))          // 현장
			router.Mount("/worker", route.WorkerRoute(safeDb, cfg, &r))                      // 근로자
			router.Mount("/compare", route.CompareRoute(safeDb, &r))                         // 일일 근로자 비교
			router.Mount("/deadline", route.DeadlineRoute(safeDb, &r))                       // 일일마감
			router.Mount("/equip", route.EquipRoute(safeDb, &r))                             // 장비 (임시)
			router.Mount("/device", route.DeviceRoute(safeDb, &r))                           // 근태인식기
			router.Mount("/company", route.CompanyRoute(safeDb, timesheetDb, &r))            // 협력업체
			router.Mount("/schedule", route.ScheduleRoute(safeDb, &r))                       // 일정관리
			router.Mount("/notice", route.NoticeRoute(safeDb, &r))                           // 공지사항
			router.Mount("/code", route.CodeRoute(safeDb, &r))                               // 코드
			router.Mount("/project-setting", route.ProjectSettingRoute(safeDb, &r))          // 프로젝트 설정
			router.Mount("/user-role", route.UserRoleRoute(safeDb, &r))                      // 사용자 권한
			router.Mount("/system", route.SystemRoute(safeDb, timesheetDb, apiCfg, cfg, &r)) // 시스템관리
		})
	})

	return c.Handler(mux), nil
}
