package service

import (
	"context"
	"github.com/go-worker-credit/internal/core"
	"github.com/go-worker-credit/internal/erro"
	"github.com/aws/aws-xray-sdk-go/xray"

)

func (s WorkerService) CreditFundSchedule(ctx context.Context, transfer core.Transfer) (error){
	childLogger.Debug().Msg("CreditFundSchedule")

	childLogger.Debug().Interface("===>transfer:",transfer).Msg("")
	
	_, root := xray.BeginSubsegment(ctx, "Service.CreditFundSchedule")
	defer root.Close(nil)

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
	}()

	// Post
	credit := core.AccountStatement{}
	credit.AccountID = transfer.AccountIDTo
	credit.Currency = transfer.Currency
	credit.Amount = transfer.Amount
	credit.Type = "CREDIT"
	
	_, err = s.restapi.PostData(ctx, s.restapi.ServerUrlDomain, s.restapi.XApigwId ,"/add", credit)
	if err != nil {
		return err
	}

	transfer.Status = "CREDIT_DONE"
	res_update, err := s.workerRepository.Update(ctx,tx ,transfer)
	if err != nil {
		return err
	}
	if res_update == 0 {
		err = erro.ErrUpdate
		return  err
	}

	return nil
}
