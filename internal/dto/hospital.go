package dto

type Hospital struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Class        string `json:"class"`
	Ownership    string `json:"ownership"`
	Address      string `json:"address"`
	PostalCode   string `json:"postal_code"`
	Phone        string `json:"phone"`
	BedsTotal    int    `json:"beds_total"`
	ICUBeds      int    `json:"icu_beds"`
	ProvinceName string `json:"province_name"`
	RegencyName  string `json:"regency_name"`
	IsActive     bool   `json:"is_active"`
}
