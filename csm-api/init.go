package main

import (
	"context"
	"csm-api/auth"
	"csm-api/clock"
	"csm-api/config"
	"csm-api/entity"
	"csm-api/service"
	"csm-api/store"
	"csm-api/utils"
	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
	"log"
)

/**
 * @author 작성자: 김진우
 * @created 작성일: 2025-04-28
 * @modified 최종 수정일:
 * @modifiedBy 최종 수정자:
 * @description: 서버 실행시 초기화를 위한 설정
 */
type Init struct {
	WorkerService   service.WorkerService
	WorkHourService service.WorkHourService
	WeatherService  service.WeatherApiService
}

func NewInit(safeDb *sqlx.DB) (*Init, error) {
	r := store.Repository{Clocker: clock.RealClock{}}
	apiCfg, err := config.GetApiConfig()
	if err != nil {
		_ = entity.WriteErrorLog(context.Background(), utils.CustomMessageErrorf("config.ApiConfig", err))
	}

	cfg, err := config.NewConfig()
	if err != nil {
		_ = entity.WriteErrorLog(context.Background(), utils.CustomMessageErrorf("config.Config", err))
	}

	init := &Init{
		WorkerService: &service.ServiceWorker{
			SafeDB:  safeDb,
			SafeTDB: safeDb,
			Store:   &r,
			Config:  cfg,
		},
		WorkHourService: &service.ServiceWorkHour{
			SafeDB:  safeDb,
			SafeTDB: safeDb,
			Store:   &r,
		},
		WeatherService: &service.ServiceWeather{
			ApiKey:       apiCfg,
			SafeDB:       safeDb,
			SafeTDB:      safeDb,
			Store:        &r,
			SitePosStore: &r,
		},
	}

	return init, nil
}

func (i *Init) RunInitializations(ctx context.Context) (err error) {
	eg, ctx := errgroup.WithContext(ctx)
	auth.SetContext(ctx, auth.UserId{}, "SYSTEM_INIT")
	auth.SetContext(ctx, auth.Uno{}, "0")

	eg.Go(func() error {
		// 현장 근로자 마감처리 (당일 이전 날짜 중에서 퇴근을 한 근로자들만 마감처리)
		// 필요시 주석 제거
		//if initErr := i.WorkerService.ModifyWorkerDeadlineInit(ctx); initErr != nil {
		//	return fmt.Errorf("[init] ModifyWorkerDeadlineInit fail: %w", initErr)
		//}
		//log.Println("[init] ModifyWorkerDeadlineInit completed")
		return nil
	})

	eg.Go(func() error {
		// 현장 근로자 공수계산 (당일 이전 날짜 중에서 출퇴퇴근을 데이터가 모드 있는 근로자만 처리)
		defer Recover("[init] ModifyWorkHour")
		log.Println("[init] ModifyWorkHour start")
		user := entity.Base{
			ModUser: utils.ParseNullString("SYSTEM_INIT"),
		}
		if initErr := i.WorkHourService.ModifyWorkHour(ctx, user); initErr != nil {
			return entity.WriteErrorLog(ctx, utils.CustomErrorf(initErr))
		}
		log.Println("[init] ModifyWorkHour completed")
		return nil
	})

	//eg.Go(func() error {
	//	defer Recover("[init] SaveWeather")
	//	log.Println("[init] SaveWeather start")
	//	err = i.WeatherService.SaveWeather(ctx)
	//	if err != nil {
	//		log.Printf("[init] SaveWeather fail: %w", err)
	//		return entity.WriteErrorLog(ctx, utils.CustomErrorf(err))
	//	}
	//	log.Printf("[init] SaveWeather completed")
	//	return nil
	//})

	//eg.Go(func() error {
	//	// 홍채인식기 전체근로자 반영
	//	defer Recover("[init] MergeRecdWorker")
	//	log.Println("[init] MergeRecdWorker start")
	//	if initErr := i.WorkerService.MergeRecdWorker(ctx); initErr != nil {
	//		return entity.WriteErrorLog(ctx, utils.CustomErrorf(initErr))
	//	}
	//	log.Println("[init] MergeRecdWorker completed")
	//	return nil
	//})
	//
	//eg.Go(func() error {
	//	// 홍채인식기 현장근로자 반영
	//	defer Recover("[init] MergeRecdDailyWorker")
	//	log.Println("[init] MergeRecdDailyWorker start")
	//	if initErr := i.WorkerService.MergeRecdDailyWorker(ctx); initErr != nil {
	//		return entity.WriteErrorLog(ctx, utils.CustomErrorf(initErr))
	//	}
	//	log.Println("[init] MergeRecdDailyWorker completed")
	//	return nil
	//})

	if err = eg.Wait(); err != nil {
		return entity.WriteErrorLog(ctx, err)
	}
	return
}
