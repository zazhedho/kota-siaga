package dto

type Earthquake struct {
	ID        string  `json:"id"`
	DateTime  string  `json:"date_time"`
	Magnitude float64 `json:"magnitude"`
	DepthKM   float64 `json:"depth_km"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Region    string  `json:"region"`
	Potential string  `json:"potential"`
	IsFelt    bool    `json:"is_felt"`
	FeltAreas *string `json:"felt_areas"`
	Source    string  `json:"source"`
}
