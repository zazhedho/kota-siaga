package dto

type Province struct {
	ID            string  `json:"id"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	AlternateName string  `json:"alternate_name"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	IsActive      bool    `json:"is_active"`
}

type City struct {
	ID            string  `json:"id"`
	ProvinceID    string  `json:"province_id"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	AlternateName string  `json:"alternate_name"`
	IsCity        bool    `json:"is_city"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	IsActive      bool    `json:"is_active"`
}

type District struct {
	ID            string  `json:"id"`
	RegencyID     string  `json:"regency_id"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	AlternateName string  `json:"alternate_name"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	IsActive      bool    `json:"is_active"`
}

type Village struct {
	ID               string  `json:"id"`
	DistrictID       string  `json:"district_id"`
	Code             string  `json:"code"`
	Name             string  `json:"name"`
	AlternateName    string  `json:"alternate_name"`
	PostalCode       string  `json:"postal_code"`
	IsCourierSupport bool    `json:"is_courier_support"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	IsActive         bool    `json:"is_active"`
}

type LocationRecord struct {
	Code       string `json:"code"`
	FullCode   string `json:"full_code"`
	Name       string `json:"name"`
	Level      string `json:"level"`
	ParentCode string `json:"parent_code"`
	PostalCode string `json:"postal_code"`
}

type LocationSearchItem struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Level      string `json:"level"`
	PostalCode string `json:"postal_code"`
	Hierarchy  string `json:"hierarchy"`
}

type LocationPath struct {
	Province Province `json:"province"`
	City     City     `json:"city"`
	District District `json:"district"`
	Level    string   `json:"level"`
	Village  *Village `json:"village,omitempty"`
}
