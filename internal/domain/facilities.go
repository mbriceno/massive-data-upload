package domain

type Facility struct {
	ID                uint         `gorm:"primaryKey"`
	FacilityName      string       `gorm:"type:varchar(150);not null"`
	AdminEntity3ID    uint         `gorm:"column:admin_entity3_id;not null"`
	AdminEntity3      AdminEntity3 `gorm:"foreignKey:AdminEntity3ID"`
	FacilityLatitude  float64      `gorm:"type:numeric;not null"`
	FacilityLongitude float64      `gorm:"type:numeric;not null"`
	CubicCapacity     int32        `gorm:"type:integer;not null"`
	FacilityType      string       `gorm:"type:varchar(250);not null"`
	IsPopupFacility   bool         `gorm:"type:boolean;default:false"`
	MainWarehouse     bool         `gorm:"type:boolean;default:false"`
	IsWarehouse       bool         `gorm:"type:boolean;default:false"`
}

func (Facility) TableName() string {
	return "tb_facilities"
}
