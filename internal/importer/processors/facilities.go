package processors

import (
	"errors"
	"massive-data-upload/internal/domain"
	"strconv"

	"gorm.io/gorm"
)

type FacilitiesProcessor struct {
	db *gorm.DB
	// Caché local para evitar ir a la DB por cada fila si los nombres se repiten
	cacheAdminEntity3 map[string]uint
}

func NewFacilitiesProcessor(db *gorm.DB) *FacilitiesProcessor {
	return &FacilitiesProcessor{
		db:                db,
		cacheAdminEntity3: make(map[string]uint),
	}
}

func (p *FacilitiesProcessor) ProcessRow(row []string) (interface{}, error) {
	if len(row) < 9 {
		return nil, errors.New("Row with insufficient columns")
	}

	facilityName := row[4]
	entity3Name := row[3]
	entity2Name := row[2]
	entity1Name := row[1]

	// Validar datos mínimos
	if facilityName == "" || entity3Name == "" || entity2Name == "" || entity1Name == "" {
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

func (p *FacilitiesProcessor) getAdminEntity3ID(entity3Name, entity2Name, entity1Name string) uint, error {
	cacheKey := entity1Name + "|" + entity2Name + "|" + entity3Name
	var entity3ID uint
	if id, exist := p.cacheAdminEntity3[cacheKey]; exist {
		entity3ID = id
	} else {
		var ae3 domain.AdminEntity3
		err := p.db.
			Joins("AdminEntity2.AdminEntity1").
			Where("tb_admin_entities3.admin_entity3_name = ?", entity3Name).
			Where("AdminEntity2.admin_entity2_name = ?", entity2Name).
			Where("AdminEntity2__AdminEntity1.admin_entity1_name = ?", entity1Name).
			First(&ae3).Error
		if err != nil {
			return 0, errors.New("Admin entity 3 not found: " + entity3Name)
		}
		entity3ID = ae3.ID
		p.cacheAdminEntity3[cacheKey] = ae3.ID
	}
	return entity3ID, nil
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
