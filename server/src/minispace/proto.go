package minispace

type ProtoAddUser struct {
	Id int `json:"id"`
	Name string `json:"name"`
}

type ProtoStopBeam struct {
	Id int `json:"id"`
	BeamId int `json:"beamid"`
	Hit int `json:"hit"`
}

type ProtoShootBeam struct {
	ShipStatus
	BeamId float64 `json:"beamid"`
}

