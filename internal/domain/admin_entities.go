package domain

import "gorm.io/gorm"

type AdminEntity1 struct {
	gorm.Model
	AdminEntity1Name string `gorm:"column:admin_entity1_name;type:varchar(150);not null"`
}

func (AdminEntity1) TableName() string {
	return "tb_admin_entities1"
}

type AdminEntity2 struct {
	gorm.Model
	AdminEntity2Name string       `gorm:"column:admin_entity2_name;type:varchar(150);not null"`
	AdminEntity1ID   uint         `gorm:"column:admin_entity1_id;not null"`
	AdminEntity1     AdminEntity1 `gorm:"foreignKey:AdminEntity1ID"`
}

func (AdminEntity2) TableName() string {
	return "tb_admin_entities2"
}

type AdminEntity3 struct {
	gorm.Model
	AdminEntity3Name string       `gorm:"column:admin_entity3_name;type:varchar(150);not null"`
	AdminEntity2ID   uint         `gorm:"column:admin_entity2_id;not null"`
	AdminEntity2     AdminEntity2 `gorm:"foreignKey:AdminEntity2ID"`
}

func (AdminEntity3) TableName() string {
	return "tb_admin_entities3"
}
