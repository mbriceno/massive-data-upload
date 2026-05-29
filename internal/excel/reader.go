package excel

import (
	"context"
	"log"
	"massive-data-upload/internal/importer"

	"github.com/xuri/excelize/v2"
)

type ReaderEngine struct {
	processors map[string]importer.TabProcessor
	batchSize  int
}

func NewReaderEngine(processors map[string]importer.TabProcessor, batchSize int) *ReaderEngine {
	return &ReaderEngine{
		processors: processors,
		batchSize:  batchSize,
	}
}

// Lee fila por fila de las pestañas, las convierte en modelos de la DB y las envía al canal
func (re *ReaderEngine) ProcessTabByStreaming(ctx context.Context, filePath string, batchChannel chan<- importer.BatchData) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Printf("[Engine] Error opening file: %v", err)
		return
	}
	defer f.Close()

	for _, sheetName := range f.GetSheetList() {
		proc, exist := re.processors[sheetName]
		if !exist {
			continue // Saltamos pestañas no mapeadas
		}

		rows, err := f.Rows(sheetName)
		if err != nil {
			log.Printf("[Engine] Error getting rows from sheet %s: %v", sheetName, err)
			continue
		}

		var bufferLote []any
		isFirstRow := true

		for rows.Next() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			columnValues, err := rows.Columns()
			if err != nil {
				continue
			}

			if isFirstRow {
				isFirstRow = false
				continue
			}

			// Validamos y mapeamos la fila individualmente en el hilo del lector
			mappedObject, err := proc.ProcessRow(columnValues)
			if err != nil {
				log.Printf("[Validation Error] %s -> %v", sheetName, err)
				continue
			}

			// Acumulamos en el buffer local de esta pestaña
			bufferLote = append(bufferLote, mappedObject)

			// Si alcanzamos el tamaño del lote, lo enviamos al pool
			if len(bufferLote) >= re.batchSize {
				batchChannel <- importer.BatchData{
					TabName:  sheetName,
					DataRows: bufferLote,
				}
				bufferLote = make([]any, 0, re.batchSize)
			}
		}

		if len(bufferLote) > 0 {
			batchChannel <- importer.BatchData{
				TabName:  sheetName,
				DataRows: bufferLote,
			}
			bufferLote = make([]any, 0, re.batchSize)
		}
	}
}
