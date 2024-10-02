package main

import(
	"sync"
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-worker-credit/internal/util"
	"github.com/go-worker-credit/internal/adapter/event"
	"github.com/go-worker-credit/internal/adapter/event/kafka"
	"github.com/go-worker-credit/internal/adapter/event/sqs"
	"github.com/go-worker-credit/internal/core"
	"github.com/go-worker-credit/internal/service"
	"github.com/go-worker-credit/internal/repository/pg"
	"github.com/go-worker-credit/internal/repository/storage"
	"github.com/go-worker-credit/internal/adapter/restapi"
)

var(
	logLevel 	= 	zerolog.DebugLevel
	appServer	core.WorkerAppServer
	consumerWorker	event.EventNotifier
)

func init() {
	log.Debug().Msg("init")
	zerolog.SetGlobalLevel(logLevel)

	infoPod, restEndpoint := util.GetInfoPod()
	database := util.GetDatabaseEnv()
	configOTEL := util.GetOtelEnv()
	kafkaConfig := util.GetKafkaEnv()
	queueConfig := util.GetQueueEnv()

	appServer.InfoPod = &infoPod
	appServer.Database = &database
	appServer.RestEndpoint = &restEndpoint
	appServer.ConfigOTEL = &configOTEL
	appServer.KafkaConfig = &kafkaConfig
	appServer.QueueConfig = &queueConfig
}

func main()  {
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Msg("main")
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Interface("appServer :",appServer).Msg("")
	log.Debug().Msg("----------------------------------------------------")

	ctx := context.Background()

	// Open Database
	count := 1
	var databasePG	pg.DatabasePG
	var err error
	for {
		databasePG, err = pg.NewDatabasePGServer(ctx, appServer.Database)
		if err != nil {
			if count < 3 {
				log.Error().Err(err).Msg("Erro open Database... trying again !!")
			} else {
				log.Error().Err(err).Msg("Fatal erro open Database aborting")
				panic(err)
			}
			time.Sleep(3 * time.Second)
			count = count + 1
			continue
		}
		break
	}

	repoDB := storage.NewWorkerRepository(databasePG)

	restApiService	:= restapi.NewRestApiService(&appServer)

	workerService := service.NewWorkerService(	&repoDB, &appServer, restApiService)

	if (appServer.InfoPod.QueueType == "kafka") {
		consumerWorker, err = kafka.NewConsumerWorker(appServer.KafkaConfig, workerService)
	} else {
		consumerWorker, err = sqs.NewNotifierSQS(ctx, appServer.QueueConfig, workerService)
	}
	if err != nil {
		log.Error().Err(err).Msg("erro connect to queue")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go consumerWorker.Consumer(	ctx, 
								&wg, 
								appServer)
	wg.Wait()
}