package importer

import "gorm.io/gorm"

type RowData struct {
	TabName string
	Data    []string
}

type TabProcessor interface {
	ProcessRow(row []string) (any, error)      // Valida y transforma la fila a un modelo
	FlushBatch(db *gorm.DB, batch []any) error // Sabe cómo insertar su lote específico
}
