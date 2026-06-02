package processors

import (
	"errors"
	"massive-data-upload/internal/domain"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DemographicsProcessor struct {
	BaseProcessor
	cacheDemographic    map[string]uint
	cachePersonsSection map[string]uint
}

func NewDemographicsProcessor(db *gorm.DB) *DemographicsProcessor {
	return &DemographicsProcessor{
		BaseProcessor: BaseProcessor{
			db:                db,
			cacheAdminEntity3: make(map[string]uint),
		},
		cacheDemographic:    make(map[string]uint),
		cachePersonsSection: make(map[string]uint),
	}
}

func (p *DemographicsProcessor) ProcessRow(row []string) (any, error) {
	if len(row) < 10 {
		return nil, errors.New("Row with insufficient columns")
	}

	entity3Name := strings.TrimSpace(row[2])
	entity2Name := strings.TrimSpace(row[1])
	entity1Name := strings.TrimSpace(row[0])

	// Validar datos mínimos
	if entity3Name == "" ||
		entity2Name == "" ||
		entity1Name == "" {
		return nil, errors.New("Facility or admin entity names are required")
	}

	entity3ID, err := p.getAdminEntity3ID(entity3Name, entity2Name, entity1Name)
	if err != nil {
		return nil, err
	}

	// 1. Validate and Convert Data
	eduInt, _ := strconv.ParseInt(strings.TrimSpace(row[3]), 10, 32)
	popInt, _ := strconv.ParseInt(strings.TrimSpace(row[4]), 10, 32)
	distInt, _ := strconv.ParseInt(strings.TrimSpace(row[5]), 10, 32)
	poverty, _ := strconv.ParseFloat(strings.TrimSpace(row[6]), 64)

	education := domain.Education(eduInt)
	population := int32(popInt)
	nearHospitalDistance := int32(distInt)

	switch education {
	case domain.EducationLow, domain.EducationMiddle, domain.EducationHigh:
		// Valid
	default:
		return nil, errors.New("invalid education choice: " + row[3])
	}

	// 2. Check if Demographic already exists to get ID
	demographicID, err := p.getDemographic(entity3ID, education, population, nearHospitalDistance, poverty)
	if err != nil {
		return nil, err
	}

	// 3. Validate Gender
	gender := domain.Gender(strings.TrimSpace(row[7]))
	if gender != domain.GenderMale && gender != domain.GenderFemale {
		return nil, errors.New("invalid gender choice: " + row[7])
	}
	age, _ := strconv.ParseInt(strings.TrimSpace(row[8]), 10, 32)
	personsNumber, _ := strconv.ParseInt(strings.TrimSpace(row[9]), 10, 32)

	// 4. Check if PersonsSection already exists
	sectionID, err := p.getPersonsSectionID(gender, int32(age), int32(personsNumber))
	if err != nil {
		return nil, err
	}

	return domain.Demographic{
		ID:                   demographicID,
		AdminEntity3ID:       entity3ID,
		Education:            education,
		Population:           population,
		NearHospitalDistance: nearHospitalDistance,
		Poverty:              poverty,
		PersonsSections: []domain.PersonsSection{
			{
				ID:            sectionID,
				Gender:        gender,
				Age:           int32(age),
				PersonsNumber: int32(personsNumber),
			},
		},
	}, nil
}

func (p *DemographicsProcessor) FlushBatch(db *gorm.DB, batch []any) error {
	demographics := make([]domain.Demographic, len(batch))
	for i, item := range batch {
		demographics[i] = item.(domain.Demographic)
	}

	// 1. Deduplicate Demographics and merge their PersonsSections
	// This prevents "ON CONFLICT DO UPDATE command cannot affect row a second time"
	type demoKey struct {
		ae3  uint
		edu  domain.Education
		pop  int32
		dist int32
		pov  float64
	}

	uniqueDemosMap := make(map[demoKey]*domain.Demographic)
	for i := range demographics {
		d := &demographics[i]
		key := demoKey{d.AdminEntity3ID, d.Education, d.Population, d.NearHospitalDistance, d.Poverty}

		if existing, ok := uniqueDemosMap[key]; ok {
			existing.PersonsSections = append(existing.PersonsSections, d.PersonsSections...)
		} else {
			uniqueDemosMap[key] = d
		}
	}

	// Reconstruct the slice with unique demographics
	uniqueDemographics := make([]domain.Demographic, 0, len(uniqueDemosMap))
	for _, d := range uniqueDemosMap {
		uniqueDemographics = append(uniqueDemographics, *d)
	}

	// 2. Deduplicate PersonsSections across all demographics in this unique set
	type sectionKey struct {
		gender domain.Gender
		age    int32
		num    int32
	}

	sectionMap := make(map[sectionKey]*domain.PersonsSection)
	for i := range uniqueDemographics {
		for j := range uniqueDemographics[i].PersonsSections {
			s := &uniqueDemographics[i].PersonsSections[j]
			key := sectionKey{s.Gender, s.Age, s.PersonsNumber}
			if existing, ok := sectionMap[key]; ok {
				uniqueDemographics[i].PersonsSections[j] = *existing
			} else {
				sectionMap[key] = s
			}
		}
	}

	// 3. Pre-save all unique PersonsSections to ensure they have IDs
	if len(sectionMap) > 0 {
		uniqueSections := make([]domain.PersonsSection, 0, len(sectionMap))
		for _, s := range sectionMap {
			uniqueSections = append(uniqueSections, *s)
		}

		// We use a dummy update to force PostgreSQL to return the ID even if the record exists
		err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "gender"}, {Name: "age"}, {Name: "persons_number"}},
			DoUpdates: clause.AssignmentColumns([]string{"gender"}),
		}).Create(&uniqueSections).Error
		if err != nil {
			return err
		}

		// Note: In a production scenario, you might need to re-map IDs back to the
		// demographics slice if GORM doesn't update the underlying pointers automatically.
	}

	// 3. Save Demographics
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "admin_entity3_id"},
			{Name: "education"},
			{Name: "population"},
			{Name: "near_hospital_distance"},
			{Name: "poverty"},
		},
		// Use a dummy update to ensure RETURNING id works for existing records
		DoUpdates: clause.AssignmentColumns([]string{"education"}),
	}).Create(&uniqueDemographics).Error
}

func (p *DemographicsProcessor) getDemographic(entity3ID uint, education domain.Education, population int32, nearHospitalDistance int32, poverty float64) (uint, error) {
	cacheKey := strconv.FormatUint(uint64(entity3ID), 10) + "|" +
		strconv.Itoa(int(education)) + "|" +
		strconv.Itoa(int(population)) + "|" +
		strconv.Itoa(int(nearHospitalDistance)) + "|" +
		strconv.FormatFloat(poverty, 'f', -1, 64)
	var demographicID uint
	if id, exist := p.cacheDemographic[cacheKey]; exist {
		demographicID = id
	} else {
		var demographics []domain.Demographic
		err := p.db.
			Where("admin_entity3_id = ?", entity3ID).
			Where("education = ?", education).
			Where("population = ?", population).
			Where("near_hospital_distance = ?", nearHospitalDistance).
			Where("poverty = ?", poverty).
			Limit(1).
			Find(&demographics).Error

		if err != nil {
			return 0, err
		}

		if len(demographics) == 0 {
			return 0, nil
		}

		demographicID = demographics[0].ID
		p.cacheDemographic[cacheKey] = demographicID
	}
	return demographicID, nil
}

func (p *DemographicsProcessor) getPersonsSectionID(gender domain.Gender, age int32, personsNumber int32) (uint, error) {
	cacheKey := string(gender) + "|" +
		strconv.Itoa(int(age)) + "|" +
		strconv.Itoa(int(personsNumber))

	if id, exist := p.cachePersonsSection[cacheKey]; exist {
		return id, nil
	}

	var sections []domain.PersonsSection
	err := p.db.
		Where("gender = ?", gender).
		Where("age = ?", age).
		Where("persons_number = ?", personsNumber).
		Limit(1).
		Find(&sections).Error

	if err != nil {
		return 0, err
	}

	if len(sections) == 0 {
		return 0, nil
	}

	p.cachePersonsSection[cacheKey] = sections[0].ID
	return sections[0].ID, nil
}
