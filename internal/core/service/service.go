package service

import(
	"github.com/go-worker-credit/internal/core/model"
	"github.com/go-worker-credit/internal/adapter/database"

	"github.com/rs/zerolog/log"
)

var childLogger = log.With().Str("component","go-worker-credit").Str("package","internal.core.service").Logger()

type WorkerService struct {
	workerRepository *database.WorkerRepository
	apiService		[]model.ApiService
}

func NewWorkerService(	workerRepository *database.WorkerRepository, 
						apiService	[]model.ApiService) *WorkerService{
	childLogger.Debug().Str("func","NewWorkerService").Send()

	return &WorkerService{
		workerRepository: workerRepository,
		apiService: apiService,
	}
}