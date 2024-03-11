package service

import (
	"context"
	//"errors"
	"github.com/go-worker-credit/internal/core"
	"github.com/go-worker-credit/internal/erro"
	"github.com/aws/aws-xray-sdk-go/xray"

)

func (s WorkerService) CreditFundSchedule(ctx context.Context, transfer core.Transfer) (error){
	childLogger.Debug().Msg("CreditFundSchedule")
	childLogger.Debug().Interface("===>transfer: ",transfer).Msg("")
	
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
	credit.Type = transfer.Type
	transfer.Status = "CREDIT_DONE"

	_, err = s.restapi.PostData(ctx, s.restapi.ServerUrlDomain, s.restapi.XApigwId ,"/add", credit)
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
