package service

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/go-worker-credit/internal/repository/postgre"
	"github.com/go-worker-credit/internal/adapter/restapi"
	"github.com/go-worker-credit/internal/core"
	"github.com/go-worker-credit/internal/erro"
	"github.com/go-worker-credit/internal/lib"
)

var childLogger = log.With().Str("service", "service").Logger()

type WorkerService struct {
	workerRepository 		*postgre.WorkerRepository
	restEndpoint			*core.RestEndpoint
	restApiService			*restapi.RestApiService
}

func NewWorkerService(	workerRepository *postgre.WorkerRepository,
						restEndpoint		*core.RestEndpoint,
						restApiService		*restapi.RestApiService) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepository:	workerRepository,
		restEndpoint:		restEndpoint,
		restApiService:		restApiService,
	}
}

func (s WorkerService) CreditFundSchedule(ctx context.Context, transfer core.Transfer) (error){
	childLogger.Debug().Msg("CreditFundSchedule")
	childLogger.Debug().Interface("===>transfer: ",transfer).Msg("")
	
	span := lib.Span(ctx, "service.CreditFundSchedule")

	tx, err := s.workerRepository.StartTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
		span.End()
	}()

	// Post
	credit := core.AccountStatement{}
	credit.AccountID = transfer.AccountIDTo
	credit.Currency = transfer.Currency
	credit.Amount = transfer.Amount
	credit.Type = transfer.Type
	transfer.Status = "CREDIT_DONE"

	urlDomain := s.restEndpoint.ServiceUrlDomain + "/add"
	_, err = s.restApiService.PostData(ctx, 
										urlDomain, 
										s.restEndpoint.XApigwId,
										credit)
	if err != nil {
		switch err{
			case erro.ErrTransInvalid:
				transfer.Status = "CREDIT_FAIL_MISMATCH_DATA"
			default:
				transfer.Status = "CREDIT_FAIL_OUTAGE"
			}
	}

	childLogger.Debug().Interface("== 2 ==> transfer update:",transfer).Msg("")
	res_update, err := s.workerRepository.Update(ctx,tx ,transfer)
	if err != nil {
		return err
	}
	if res_update == 0 {
		err = erro.ErrUpdate
		return err
	}

	if transfer.Status != "CREDIT_DONE"{
		return erro.ErrEvent
	}

	return nil
}