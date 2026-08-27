package model

import pkgerrs "github.com/maket12/ads-service/backend/authservice/pkg/errs"

type Category string

const (
	CategoryElectronics Category = "electronics"
	CategoryVehicles    Category = "vehicles"
	CategoryRealEstate  Category = "real_estate"
	CategoryClothes     Category = "clothes"
	CategoryFood        Category = "food"
	CategoryHome        Category = "home"
	CategoryServices    Category = "services"
	CategoryVideoGames  Category = "video_games"
	CategoryMedicaments Category = "medicaments"
	CategoryTravel      Category = "travel"
)

func (c Category) String() string { return string(c) }

func (c Category) IsValid() bool {
	switch c {
	case CategoryElectronics, CategoryVehicles,
		CategoryRealEstate, CategoryClothes,
		CategoryFood, CategoryHome,
		CategoryServices, CategoryVideoGames,
		CategoryMedicaments, CategoryTravel:
		return true
	default:
		return false
	}
}

func NewCategory(val string) (Category, error) {
	cat := Category(val)
	if !cat.IsValid() {
		return "", pkgerrs.NewValueInvalidError("category")
	}
	return cat, nil
}
