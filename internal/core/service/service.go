package service

import(
	"github.com/go-worker-credit/internal/core/model"
	"github.com/go-worker-credit/internal/adapter/database"

	"github.com/rs/zerolog/log"
)

var childLogger = log.With().Str("core", "service").Logger()

type WorkerService struct {
	workerRepository *database.WorkerRepository
	apiService		[]model.ApiService
}

func NewWorkerService(	workerRepository *database.WorkerRepository, 
						apiService	[]model.ApiService) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepository: workerRepository,
		apiService: apiService,
	}
}