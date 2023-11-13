package model

import "time"

type Local struct {
	ID                 int       `json:"id" csv:"id"`
	Latitude           float64   `json:"latitude" csv:"latitude"`
	Longitude          float64   `json:"longitude" csv:"longitude"`
	AvailabilityRadius int       `json:"availability_radius" csv:"availability_radius"`
	OpenHour           time.Time `json:"open_hour" csv:"open_hour"`
	CloseHour          time.Time `json:"close_hour" csv:"close_hour"`
	Rating             float64   `json:"rating" csv:"rating"`
}
