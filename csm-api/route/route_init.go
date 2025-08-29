package route

import (
	"csm-api/config"
	"csm-api/handler"
	"csm-api/service"
	"csm-api/store"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

func InitApiRoute(safeDb *sqlx.DB, timesheetDb *sqlx.DB, apiCfg *config.ApiConfig, cfg *config.Config, r *store.Repository) chi.Router {
	router := chi.NewRouter()

	initApihandler := &handler.InitApiHandler{

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

	router.Get("/", initApihandler.ServeHTTP)

	return router
}
