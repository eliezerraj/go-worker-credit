package event

import (
	"context"
	"sync"
	
	"github.com/go-worker-credit/internal/core"
)

type EventNotifier interface {
	Consumer(ctx context.Context, wg *sync.WaitGroup, appServer core.WorkerAppServer)
}