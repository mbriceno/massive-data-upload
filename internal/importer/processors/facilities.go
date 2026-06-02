package processors

import (
	"errors"
	"massive-data-upload/internal/domain"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type FacilitiesProcessor struct {
	BaseProcessor
}

func NewFacilitiesProcessor(db *gorm.DB) *FacilitiesProcessor {
	return &FacilitiesProcessor{
		BaseProcessor: BaseProcessor{
			db:                db,
			cacheAdminEntity3: make(map[string]uint),
		},
	}
}

func (p *FacilitiesProcessor) ProcessRow(row []string) (any, error) {
	if len(row) < 9 {
		return nil, errors.New("Row with insufficient columns")
	}

	facilityName := strings.TrimSpace(row[4])
	entity3Name := strings.TrimSpace(row[3])
	entity2Name := strings.TrimSpace(row[2])
	entity1Name := strings.TrimSpace(row[1])

	// Validar datos mínimos
	if facilityName == "" ||
		entity3Name == "" ||
		entity2Name == "" ||
		entity1Name == "" {
		return nil, errors.New("Facility or admin entity names are required")
	}

	// Buscar el ID del AdminEntity3 (Uso de caché transitoria)

	entity3ID, err := p.getAdminEntity3ID(entity3Name, entity2Name, entity1Name)
	if err != nil {
		return nil, err
	}

	isWarehouse, _ := strconv.ParseBool(row[5])
	mainWarehouse, _ := strconv.ParseBool(row[6])
	lat, _ := strconv.ParseFloat(row[7], 64)
	long, _ := strconv.ParseFloat(row[8], 64)
	capacity, _ := strconv.ParseInt(row[9], 10, 32)
	isPopup, _ := strconv.ParseBool(row[10])

	return domain.Facility{
		AdminEntity3ID:    entity3ID,
		FacilityName:      facilityName,
		FacilityType:      row[0],
		IsWarehouse:       isWarehouse,
		MainWarehouse:     mainWarehouse,
		FacilityLatitude:  lat,
		FacilityLongitude: long,
		CubicCapacity:     int32(capacity),
		IsPopupFacility:   isPopup,
	}, nil
}

func (p *FacilitiesProcessor) FlushBatch(db *gorm.DB, batch []any) error {
	// Convertimos el slice genérico al tipo struct concreto de GORM para Bulk Insert
	facilities := make([]domain.Facility, len(batch))
	for i, item := range batch {
		facilities[i] = item.(domain.Facility)
	}
	// GORM ejecutará un único INSERT masivo y tipado
	return db.Create(&facilities).Error
}
