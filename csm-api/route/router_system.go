package route

import (
	"csm-api/config"
	"csm-api/handler"
	"csm-api/service"
	"csm-api/store"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

func SystemRoute(safeDb *sqlx.DB, timesheetDb *sqlx.DB, apiCfg *config.ApiConfig, cfg *config.Config, r *store.Repository) chi.Router {

	router := chi.NewRouter()

	systemHandler := &handler.SystemHandler{

		WorkerService: &service.ServiceWorker{
			SafeDB:  safeDb,
			SafeTDB: safeDb,
			Store:   r,
			Config:  cfg,
		},
		WorkHourService: &service.ServiceWorkHour{
			SafeDB:  safeDb,
			SafeTDB: safeDb,
			Store:   r,
		},
		ProjectService: &service.ServiceProject{
			SafeDB:  safeDb,
			SafeTDB: safeDb,
			Store:   r,
		},
		ProjectSettingService: &service.ServiceProjectSetting{
			SafeDB:        safeDb,
			SafeTDB:       safeDb,
			Store:         r,
			WorkHourStore: r,
		},
		WeatherService: &service.ServiceWeather{
			ApiKey:       apiCfg,
			SafeDB:       safeDb,
			SafeTDB:      safeDb,
			Store:        r,
			SitePosStore: r,
		},
		SiteService: &service.ServiceSite{
			SafeDB:            safeDb,
			SafeTDB:           safeDb,
			Store:             r,
			ProjectStore:      r,
			ProjectDailyStore: r,
			SitePosStore:      r,
			SiteDateStore:     r,
			UserService: &service.ServiceUser{
				SafeDB:      safeDb,
				TimeSheetDB: timesheetDb,
				Store:       r,
			},
		},
	}

	router.Get("/worker-deadline", systemHandler.WorkerDeadline)     // 근로자 마감 처리
	router.Get("/worker-overtime", systemHandler.WorkerOverTime)     // 철야 확인 작업
	router.Get("/project-setting", systemHandler.ProjectInitSetting) // 프로젝트 정보 업데이트
	router.Get("/update-workhour", systemHandler.UpdateWorkHour)     // 근로자 공수 계산
	router.Get("/setting-workrate", systemHandler.SettingWorkRate)   // 당일 공정률 기록
	router.Post("/manhour", systemHandler.AddManHour)                // 공수 추가

	return router

}
