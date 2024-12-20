package service

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/go-worker-credit/internal/repository/storage"
	"github.com/go-worker-credit/internal/adapter/restapi"
	"github.com/go-worker-credit/internal/core"
	"github.com/go-worker-credit/internal/erro"
	"github.com/go-worker-credit/internal/lib"
)

var childLogger = log.With().Str("service", "service").Logger()
var restApiCallData core.RestApiCallData

type WorkerService struct {
	workerRepo		*storage.WorkerRepository
	appServer		*core.WorkerAppServer
	restApiService	*restapi.RestApiService
}

func NewWorkerService(	workerRepo		*storage.WorkerRepository,
						appServer		*core.WorkerAppServer,
						restApiService	*restapi.RestApiService) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepo:	workerRepo,
		appServer:	appServer,
		restApiService:	restApiService,
	}
}

func (s WorkerService) CreditFundSchedule(ctx context.Context, transfer core.Transfer) (error){
	childLogger.Debug().Msg("CreditFundSchedule")
	childLogger.Debug().Interface("===>transfer: ",transfer).Msg("")
	
	span := lib.Span(ctx, "service.CreditFundSchedule")

	tx, conn, err := s.workerRepo.StartTx(ctx)
	if err != nil {
		return err
	}
	
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepo.ReleaseTx(conn)
		span.End()
	}()

	// Post
	credit := core.AccountStatement{}
	credit.AccountID = transfer.AccountIDTo
	credit.Currency = transfer.Currency
	credit.Amount = transfer.Amount
	credit.Type = transfer.Type
	transfer.Status = "CREDIT_DONE"

	restApiCallData.Method = "POST"
	restApiCallData.Url = s.appServer.RestEndpoint.ServiceUrlDomain + "/add/"
	restApiCallData.X_Api_Id = &s.appServer.RestEndpoint.XApigwId

	_, err = s.restApiService.CallApiRest(ctx, restApiCallData, credit)
	if err != nil {
		switch err{
			case erro.ErrTransInvalid:
				transfer.Status = "CREDIT_FAIL_MISMATCH_DATA"
			default:
				transfer.Status = "CREDIT_FAIL_OUTAGE"
			}
	}

	childLogger.Debug().Interface("== 2 ==> transfer update:",transfer).Msg("")
	res_update, err := s.workerRepo.Update(ctx,tx, &transfer)
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