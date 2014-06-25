package minispace

type AIAction interface {
	ActShoot() error
//	ActMove(int)
	ActRotate(int)
}

type AIAlgo interface {
	Update(ai AIAction, dt float64)
	Name() string
}

type AIUser struct {
	Ship
	name string
	act int
	algo AIAlgo
}

func (ai *AIUser) UserName() string {
	return ai.name
}

func (ai *AIUser) Die() {
	s := ai.scene

	// recover...
	ai.Hp = 100

	// remove self from dead list
	ai.sceneList.RemoveSelf()

	// push back to active list
	s.activeList.PushBack(&ai.sceneList)
}

func (ai *AIUser) Beat() {
}

func (ai *AIUser) updatePosition(delta float64) {
	if ai.Rotate == 2 {
		angle := ai.Angle + 80 * (delta / 1000);
		if angle >= 360 {
			angle = angle - 360;
		}
		ai.Angle = angle
	}
}

func (ai *AIUser) updateAction(delta float64) {
	if ai.act == 1 {
		// shoot
	}

	// clear
	ai.act = 0
}

func (ai *AIUser) Update(delta float64) {
	ai.algo.Update(ai, delta)

	ai.updateAction(delta)
	ShipUpdateBeam(ai, delta)
	ai.updatePosition(delta)
}

func (ai *AIUser) GetShip() *Ship {
	return &ai.Ship
}

func (ai *AIUser) SendClient(msg *Msg) error {
	return nil
}

// for AIAction
func (ai *AIUser) ActRotate(dir int) {
	ai.Rotate = dir
}

func (ai *AIUser) ActShoot() error {
	mask := ai.beamMap
	id := 0

	for ((mask & 0x01) > 0) && id < 5 {
		id++
		mask = mask >> 1
	}

	if id >= 5 {
		return nil
	}

	ShipAddBeam(ai, id)
	return nil
}

func NewAIUser() *AIUser {
	ai := &AIUser{
		name: "AI",
	}
	InitShip(&ai.Ship)
	ai.X = 480
	ai.Y = 320
	ai.Hp = 100
	InitList(&ai.sceneList, ai)
	ai.algo = NewAISimapleAlgo()
	return ai
}
