package route

import (
	"csm-api/config"
	"csm-api/handler"
	"csm-api/service"
	"csm-api/store"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

func SiteRoute(safeDB *sqlx.DB, timeSheetDB *sqlx.DB, r *store.Repository, apiConfig *config.ApiConfig) chi.Router {
	router := chi.NewRouter()

	siteHandler := &handler.HandlerSite{
		Service: &service.ServiceSite{
			SafeDB:            safeDB,
			SafeTDB:           safeDB,
			Store:             r,
			ProjectStore:      r,
			ProjectDailyStore: r,
			SitePosStore:      r,
			SiteDateStore:     r,
			UserService: &service.ServiceUser{
				SafeDB:      safeDB,
				TimeSheetDB: timeSheetDB,
				Store:       r,
			},
			ProjectService: &service.ServiceProject{
				SafeDB:    safeDB,
				Store:     r,
				UserStore: r,
			},
			WeatherApiService: &service.ServiceWeather{
				ApiKey:       apiConfig,
				SafeDB:       safeDB,
				SafeTDB:      safeDB,
				Store:        r,
				SitePosStore: r,
			},
			AddressSearchAPIService: &service.ServiceAddressSearch{
				ApiKey: apiConfig,
			},
			RestDateApiService: &service.ServiceRestDate{
				ApiKey: apiConfig,
			},
		},
		CodeService: &service.ServiceCode{
			SafeDB: safeDB,
			Store:  r,
		},
	}

	router.Get("/", siteHandler.List)                                      // 현장관리 조회
	router.Get("/nm", siteHandler.SiteNameList)                            // 현장명 조회
	router.Get("/stats", siteHandler.StatsList)                            // 현장상태조회
	router.Post("/", siteHandler.Add)                                      // 현장 생성
	router.Put("/", siteHandler.Modify)                                    // 수정
	router.Delete("/{sno}", siteHandler.Delete)                            // 현장 삭제
	router.Put("/non-use", siteHandler.ModifyNonUse)                       // 현장 사용안함
	router.Put("/use", siteHandler.ModifyUse)                              // 현장 사용
	router.Put("/non-use/job", siteHandler.ModifySiteJobNonUse)            // 현장 프로젝트 사용안함
	router.Put("/use/job", siteHandler.ModifySiteJobUse)                   // 현장 프로젝트 사용
	router.Put("/work-rate", siteHandler.ModifyWorkRate)                   // 공정률 수정
	router.Get("/work-rate", siteHandler.SiteWorkRateByDate)               // 날짜별 현장 공정률
	router.Get("/work-rate/{jno}/{date}", siteHandler.SiteWorkRateByMonth) // 월별 공정률
	router.Post("/work-rate", siteHandler.AddWorkRate)                     // 공정률 추가

	return router
}
