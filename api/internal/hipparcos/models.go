package hipparcos

type HipparcosHR struct {
	HIP  int     `json:"hip" bson:"HIP"`
	Amag float64 `json:"Amag" bson:"Amag"`
	BV   float64 `json:"BV" bson:"BV"`
}
