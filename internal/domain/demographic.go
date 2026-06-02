package domain

type Education int32
type Gender string

const (
	EducationLow    Education = 0
	EducationMiddle Education = 1
	EducationHigh   Education = 2
)

const (
	GenderMale   Gender = "M"
	GenderFemale Gender = "F"
)

type Demographic struct {
	ID                   uint             `gorm:"primaryKey"`
	AdminEntity3ID       uint             `gorm:"column:admin_entity3_id;not null;uniqueIndex:tb_demographics_idx"`
	AdminEntity3         AdminEntity3     `gorm:"foreignKey:AdminEntity3ID"`
	Education            Education        `gorm:"type:integer;not null;uniqueIndex:tb_demographics_idx"`
	Population           int32            `gorm:"type:integer;uniqueIndex:tb_demographics_idx"`
	NearHospitalDistance int32            `gorm:"type:integer;uniqueIndex:tb_demographics_idx"`
	Poverty              float64          `gorm:"type:numeric;uniqueIndex:tb_demographics_idx"`
	PersonsSections      []PersonsSection `gorm:"many2many:tb_demographics_persons_sections;joinForeignKey:demographic_id;joinReferences:personssection_id"`
}

func (Demographic) TableName() string {
	return "tb_demographics"
}

type PersonsSection struct {
	ID            uint   `gorm:"primaryKey"`
	Gender        Gender `gorm:"type:varchar(1);not null;uniqueIndex:idx_persons_section_unique"`
	Age           int32  `gorm:"type:integer;uniqueIndex:idx_persons_section_unique"`
	PersonsNumber int32  `gorm:"type:integer;uniqueIndex:idx_persons_section_unique"`
}

func (PersonsSection) TableName() string {
	return "tb_persons_sections"
}
