package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBClient envuelve la conexión física para exponer métodos limpios
type DBClient struct {
	GormDB *gorm.DB
}

// NewPostgresClient inicializa el pool de conexiones con optimizaciones
func NewPostgresClient(dsn string, batchSize int) (*DBClient, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		CreateBatchSize: batchSize,
	})
	if err != nil {
		return nil, err
	}

	// Optimizaciones del pool nativo
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	return &DBClient{GormDB: db}, nil
}

// InsertBatch ejecuta la persistencia real. Es la función que sacamos del worker.
func (c *DBClient) InsertBatch(tableName string, rowsData []map[string]interface{}, workerID int) error {
	if len(rowsData) == 0 {
		return nil
	}

	inicio := time.Now()
	dbTableName := fmt.Sprintf("tb_%s", tableName)
	err := c.GormDB.Table(dbTableName).Create(&rowsData).Error
	if err != nil {
		return fmt.Errorf("Error in Bulk Insert for %d rows on table %s: %w", len(rowsData), tableName, err)
	}

	fmt.Printf("💾 [Worker %d] [DB-Postgres] Bulk Insert success for %d rows in '%s' [%v]\n", workerID, len(rowsData), tableName, time.Since(inicio))
	return nil
}
