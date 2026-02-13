package models

type LocationMessage struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

func (l LocationMessage) Valid() bool {
	if l.VehicleID == "" {
		return false
	}
	if l.Latitude < -90 || l.Latitude > 90 {
		return false
	}
	if l.Longitude < -180 || l.Longitude > 180 {
		return false
	}
	if l.Timestamp <= 0 {
		return false
	}
	return true
}
