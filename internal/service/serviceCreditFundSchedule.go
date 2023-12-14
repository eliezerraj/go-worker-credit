package service

import (
	"context"
	"github.com/go-worker-credit/internal/core"
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

	_, err = s.restapi.PostData(ctx, s.restapi.ServerUrlDomain ,"/add", credit)
	if err != nil {
		return err
	}

	transfer.Status = "CREDIT_DONE"
	_, err = s.workerRepository.Update(ctx,tx ,transfer)
	if err != nil {
		return err
	}

	return nil
}
