package importer

import (
	"context"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

const (
	BatchSize   = 500
	MaxWaitTime = 2 * time.Second
)

type BatchData struct {
	TabName  string
	DataRows []any
}

type WorkerPool struct {
	db         *gorm.DB
	processors map[string]TabProcessor
}

func NewWorkerPool(db *gorm.DB, processors map[string]TabProcessor) *WorkerPool {
	return &WorkerPool{
		db:         db,
		processors: processors,
	}
}

// Start representa a uno de los hilos del pool encargado de escribir en lotes en PostgreSQL
func (wp *WorkerPool) Start(ctx context.Context, id int, channelIn <-chan BatchData, wg *sync.WaitGroup) {
	defer wg.Done()

	for batch := range channelIn {
		select {
		case <-ctx.Done():
			log.Printf("[Worker %d] Context canceled. Saving residual data...\n", id)
			return
		default:
		}

		proc, exist := wp.processors[batch.TabName]
		if !exist {
			log.Printf("[Worker %d] Processor not found for tab: %s. Skipping batch.", id, batch.TabName)
			continue
		}

		log.Printf("[Worker-%d] Inserting batch of %d rows for tab: %s", id, len(batch.DataRows), batch.TabName)

		if err := proc.FlushBatch(wp.db, batch.DataRows); err != nil {
			log.Printf("[Worker-%d][DB Error] Error in %s: %v", id, batch.TabName, err)
		}
	}

	log.Printf("[Worker-%d] Channel closed. Exiting goroutine.", id)
}
