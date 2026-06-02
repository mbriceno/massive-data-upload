package processors

import (
	"errors"
	"massive-data-upload/internal/domain"
	"strings"

	"gorm.io/gorm"
)

type BaseProcessor struct {
	db                *gorm.DB
	cacheAdminEntity3 map[string]uint
}

func (p *BaseProcessor) getAdminEntity3ID(entity3Name, entity2Name, entity1Name string) (uint, error) {
	cacheKey := strings.ToLower(entity1Name) + "|" +
		strings.ToLower(entity2Name) + "|" +
		strings.ToLower(entity3Name)

	if id, exist := p.cacheAdminEntity3[cacheKey]; exist {
		return id, nil
	}

	var ae3 domain.AdminEntity3
	err := p.db.
		Joins("AdminEntity2.AdminEntity1").
		Where("tb_admin_entities3.admin_entity3_name = ?", entity3Name).
		Where("\"AdminEntity2\".admin_entity2_name = ?", entity2Name).
		Where("\"AdminEntity2__AdminEntity1\".admin_entity1_name = ?", entity1Name).
		First(&ae3).Error
	if err != nil {
		return 0, errors.New("Admin entity 3 not found: " + entity3Name)
	}

	p.cacheAdminEntity3[cacheKey] = ae3.ID
	return ae3.ID, nil
}
