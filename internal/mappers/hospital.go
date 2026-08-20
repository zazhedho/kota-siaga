package mappers

import "kota-siaga/internal/dto"

type HospitalSource struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Jenis        string `json:"jenis"`
	Kelas        string `json:"kelas"`
	Ownership    string `json:"ownership"`
	Address      string `json:"address"`
	PostalCode   string `json:"postal_code"`
	Phone        string `json:"phone"`
	BedsTotal    int    `json:"beds_total"`
	ICUBeds      int    `json:"icu_beds"`
	ProvinceName string `json:"province_name"`
	RegencyName  string `json:"regency_name"`
	IsActive     int    `json:"is_active"`
}

func MapHospital(source HospitalSource) dto.Hospital {
	return dto.Hospital{
		ID:           source.ID,
		Name:         source.Name,
		Type:         source.Jenis,
		Class:        source.Kelas,
		Ownership:    source.Ownership,
		Address:      source.Address,
		PostalCode:   source.PostalCode,
		Phone:        source.Phone,
		BedsTotal:    source.BedsTotal,
		ICUBeds:      source.ICUBeds,
		ProvinceName: source.ProvinceName,
		RegencyName:  source.RegencyName,
		IsActive:     source.IsActive != 0,
	}
}

func MapHospitalPage(source dto.Page[HospitalSource]) dto.Page[dto.Hospital] {
	return dto.Page[dto.Hospital]{
		Data:       mapItems(source.Data, MapHospital),
		Total:      source.Total,
		Page:       source.Page,
		PerPage:    source.PerPage,
		TotalPages: source.TotalPages,
	}
}
