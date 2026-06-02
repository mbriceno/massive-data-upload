package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"massive-data-upload/internal/config"
	"massive-data-upload/internal/database"
	"massive-data-upload/internal/excel"
	"massive-data-upload/internal/importer"
	"massive-data-upload/internal/importer/processors"
)

func main() {
	// 1. Definir banderas (flags) de la línea de comandos
	excelPathFlag := flag.String("file", "", "Path massive upload file in Excel (Eg: --file=datos.xlsx)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Use for %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	// Validar obligatoriedad del archivo
	if *excelPathFlag == "" {
		fmt.Println("❌ Error: Excel file path is required.")
		flag.Usage()
		os.Exit(1)
	}

	// 2. Cargar Configuración pasando el parámetro CLI (el .env se lee adentro)
	cfg := config.Load()

	// 3. Inicializar la base de datos con las credenciales del .env
	dbClient, err := database.NewPostgresClient(cfg.DBDSN, cfg.NumWorkers)
	if err != nil {
		log.Fatalf("❌ Error starting Postgres: %v", err)
	}

	// 4. Infraestructura de canales y concurrencia
	batchChannel := make(chan importer.BatchData, 16)
	var wgWorkers sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	processorsRegistry := map[string]importer.TabProcessor{
		"facilities":   processors.NewFacilitiesProcessor(dbClient.GormDB),
		"demographics": processors.NewDemographicsProcessor(dbClient.GormDB),
	}

	pool := importer.NewWorkerPool(dbClient.GormDB, processorsRegistry)

	// 5. Encender Workers dinámicos basados en la configuración del .env
	for i := 1; i <= cfg.NumWorkers; i++ {
		wgWorkers.Add(1)
		go pool.Start(ctx, i, batchChannel, &wgWorkers)
	}

	// 6. Lectura por Streaming del archivo Excel CLI
	startTime := time.Now()
	fmt.Printf("🚀 Starting dynamic pipeline [File: %s] [Workers: %d]...\n", *excelPathFlag, cfg.NumWorkers)

	reader := excel.NewReaderEngine(processorsRegistry, importer.BatchSize)
	reader.ProcessTabByStreaming(ctx, *excelPathFlag, batchChannel)

	close(batchChannel)
	wgWorkers.Wait()

	fmt.Printf("✨ ¡Successful massive upload finished %v!\n", time.Since(startTime))
}
