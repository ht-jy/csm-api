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
	"github.com/robfig/cron/v3"
	"log"
	"time"
)

/**
 * @author 작성자: 김진우
 * @created 작성일: 2025-04-28
 * @modified 최종 수정일:
 * @modifiedBy 최종 수정자:
 * @description: 스캐줄러 설정
 * -
 */

type Scheduler struct {
	WorkerService         service.WorkerService
	WorkHourService       service.WorkHourService
	ProjectService        service.ProjectService
	ProjectSettingService service.ProjectSettingService
	WeatherService        service.WeatherApiService
	SiteService           service.SiteService
	cron                  *cron.Cron
}

func NewScheduler(safeDb *sqlx.DB, apiCfg *config.ApiConfig, cfg *config.Config, timesheetDb *sqlx.DB) (*Scheduler, error) {
	r := store.Repository{Clocker: clock.RealClock{}}
	c := cron.New(cron.WithSeconds())

	scheduler := &Scheduler{
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
		ProjectService: &service.ServiceProject{
			SafeDB:  safeDb,
			SafeTDB: safeDb,
			Store:   &r,
		},
		ProjectSettingService: &service.ServiceProjectSetting{
			SafeDB:        safeDb,
			SafeTDB:       safeDb,
			Store:         &r,
			WorkHourStore: &r,
		},
		WeatherService: &service.ServiceWeather{
			ApiKey:       apiCfg,
			SafeDB:       safeDb,
			SafeTDB:      safeDb,
			Store:        &r,
			SitePosStore: &r,
		},
		SiteService: &service.ServiceSite{
			SafeDB:            safeDb,
			SafeTDB:           safeDb,
			Store:             &r,
			ProjectStore:      &r,
			ProjectDailyStore: &r,
			SitePosStore:      &r,
			SiteDateStore:     &r,
			UserService: &service.ServiceUser{
				SafeDB:      safeDb,
				TimeSheetDB: timesheetDb,
				Store:       &r,
			},
		},

		cron: c,
	}

	return scheduler, nil
}

func (s *Scheduler) Run(ctx context.Context) error {
	auth.SetContext(ctx, auth.UserId{}, "SYSTEM_SCHEDULER")
	auth.SetContext(ctx, auth.Uno{}, "0")

	// 근로자 마감 처리 (퇴근한 근로자만 처리)::5시 0분 0초
	_, err := s.cron.AddFunc("0 0 5 * * *", func() {
		defer Recover("[Scheduler] Running ModifyWorkerDeadlineSchedule")
		log.Println("[Scheduler] Running ModifyWorkerDeadlineSchedule")
		if err := s.WorkerService.ModifyWorkerDeadlineInit(ctx); err != nil {
			log.Println("[Scheduler] ModifyWorkerDeadlineSchedule fail")
			_ = entity.WriteErrorLog(ctx, utils.CustomMessageErrorf("[Scheduler] ModifyWorkerDeadlineSchedule", err))
		} else {
			log.Println("[Scheduler] ModifyWorkerDeadlineSchedule completed")
		}
	})
	if err != nil {
		return entity.WriteErrorLog(ctx, utils.CustomMessageErrorf("[Scheduler] failed to add cron job", err))
	}

	// 철야 확인 작업::1분마다
	//_, err = s.cron.AddFunc("0 0/1 * * * *", func() {
	//	var count int
	//	if count, err = s.WorkerService.ModifyWorkerOverTime(ctx); err != nil {
	//		_ = entity.WriteErrorLog(ctx, utils.CustomMessageErrorf("[Scheduler] ModifyWorkerOverTime", err))
	//	} else if count != 0 {
	//		log.Println("[Scheduler] ModifyWorkerOverTime completed")
	//	}
	//})
	//if err != nil {
	//	return entity.WriteErrorLog(ctx, utils.CustomMessageErrorf("[Scheduler] failed to add cron job", err))
	//}

	// 프로젝트 정보 업데이트(초기 세팅)::5분
	_, err = s.cron.AddFunc("0 0/5 * * * *", func() {
		defer Recover("[Scheduler] Running CheckProjectSettings")
		//log.Println("[Scheduler] Running CheckProjectSettings")
		var count int
		if count, err = s.ProjectSettingService.CheckProjectSetting(ctx); err != nil {
			log.Printf("[Scheduler] CheckProjectSettings fail: %+v", err)
			_ = entity.WriteErrorLog(ctx, utils.CustomMessageErrorf("[Scheduler] CheckProjectSettings", err))
		} else if count != 0 {
			log.Printf("[Scheduler] CheckProjectSettings %d completed \n", count)
		}
	})
	if err != nil {
		return entity.WriteErrorLog(ctx, utils.CustomMessageErrorf("[Scheduler] failed to add cron job", err))
	}

	// 근로자 공수 계산 (마감 처리가 안되고 출퇴근이 모두 있는 근로자)::0시 1분 0초
	_, err = s.cron.AddFunc("0 1 0 * * *", func() {
		defer Recover("[Scheduler] Running ModifyWorkHour")
		log.Println("[Scheduler] Running ModifyWorkHour")
		user := entity.Base{
			ModUser: utils.ParseNullString("SYSTEM_BATCH"),
		}
		if err = s.WorkHourService.ModifyWorkHour(ctx, user); err != nil {
			log.Println("[Scheduler] ModifyWorkHour fail")
			_ = entity.WriteErrorLog(ctx, utils.CustomMessageErrorf("[Scheduler] ModifyWorkHour", err))
		} else {
			log.Println("[Scheduler] ModifyWorkHour completed")
		}
	})
	if err != nil {
		return entity.WriteErrorLog(ctx, utils.CustomMessageErrorf("[Scheduler] failed to add cron job", err))
	}

	// 날씨 저장::8시, 10시, 13시, 15시, 17시
	_, err = s.cron.AddFunc("0 0 8,10,13,15,17 * * *", func() {
		defer Recover("[Scheduler] Running SaveWeather")
		log.Println("[Scheduler] Running SaveWeather")

		err = s.WeatherService.SaveWeather(ctx)
		if err != nil {
			log.Printf("[Scheduler] SaveWeather fail: %w", err)
			_ = entity.WriteErrorLog(ctx, utils.CustomMessageErrorf("[Scheduler] SaveWeather", err))
		} else {
			log.Printf("[Scheduler] SaveWeather completed")
		}

	})
	if err != nil {
		return entity.WriteErrorLog(ctx, utils.CustomMessageErrorf("[Scheduler] failed to add cron job", err))
	}

	// 공정률 기록::00:00부터 05:00까지 1시간
	_, err = s.cron.AddFunc("0 0 0,1,2,3,4,5 * * *", func() {
		defer Recover("[Scheduler] Running SettingWorkRate")
		log.Println("[Scheduler] Running SettingWorkRate")
		var count int64
		now := time.Now()
		count, err = s.SiteService.SettingWorkRate(ctx, now)
		if err != nil {
			log.Printf("[Scheduler] SettingWorkRate fail: %w", err)
			_ = entity.WriteErrorLog(ctx, utils.CustomMessageErrorf("[Scheduler] SettingWorkRate", err))
		} else if count > 0 {
			log.Printf("[Scheduler] SettingWorkRate success: %d", count)
		}
	})
	if err != nil {
		return utils.CustomMessageErrorf("[Scheduler] failed to add cron job", err)
	}

	// 홍채인식기 전체근로자 반영::2분
	_, err = s.cron.AddFunc("0 0/2 * * * *", func() {
		defer Recover("[Scheduler] Running MergeRecdWorker")
		//log.Println("[Scheduler] Running MergeRecdWorker")
		if err = s.WorkerService.MergeRecdWorker(ctx); err != nil {
			log.Println("[Scheduler] MergeRecdWorker fail")
			_ = entity.WriteErrorLog(ctx, utils.CustomMessageErrorf("[Scheduler] MergeRecdWorker", err))
		} else {
			log.Println("[Scheduler] MergeRecdWorker completed")
		}
	})
	if err != nil {
		return entity.WriteErrorLog(ctx, utils.CustomMessageErrorf("[Scheduler] failed to add cron job", err))
	}

	// 홍채인식기 현장근로자 반영::2분(정시 1분부터 시작)
	_, err = s.cron.AddFunc("0 1-59/2 * * * *", func() {
		defer Recover("[Scheduler] Running MergeRecdDailyWorker")
		//log.Println("[Scheduler] Running MergeRecdDailyWorker")
		if err = s.WorkerService.MergeRecdDailyWorker(ctx); err != nil {
			log.Println("[Scheduler] MergeRecdDailyWorker fail")
			_ = entity.WriteErrorLog(ctx, utils.CustomMessageErrorf("[Scheduler] MergeRecdDailyWorker", err))
		} else {
			log.Println("[Scheduler] MergeRecdDailyWorker completed")
		}
	})
	if err != nil {
		return entity.WriteErrorLog(ctx, utils.CustomMessageErrorf("[Scheduler] failed to add cron job", err))
	}

	// ... 추가 job 등록
	s.cron.Start()

	log.Println("Scheduler server start")

	// ctx.Done() 기다리다가 종료
	<-ctx.Done()
	log.Println("[Scheduler] Stopping scheduler...")

	s.cron.Stop()
	log.Println("[Scheduler] Stop")
	return nil
}
