package minispace

import "fmt"
import "container/list"

type AIAlgo interface {
	Update(ai *AIUser)
	Name() string
}

type AIUser struct {
	id int
	name string
	scene *Scene
	sceneList List
	x, y, angle, hp, move, rotate, act int
	beamMap int
	beamList *list.List
	algo AIAlgo
}

func (ai *AIUser) SendClient(msg *Msg) error {
	return nil
}

func (ai *AIUser) SetScene(s *Scene) {
	ai.scene = s
}

func (ai *AIUser) SetUserId(id int) {
	ai.id = id
}

func (ai *AIUser) UserId() int {
	return ai.id
}

func (ai *AIUser) UserName() string {
	return ai.name
}

func (ai *AIUser) Position() (int, int) {
	return ai.x, ai.y
}

func (ai *AIUser) HpDown(value int) int {
	ai.hp -= value
	if ai.hp < 0 {
		ai.hp = 0
	}
	return ai.hp
}

func (ai *AIUser) Die() {
	s := ai.scene

	// recover...
	ai.hp = 100

	// remove self from dead list
	ai.sceneList.RemoveSelf()

	// push back to active list
	s.activeList.PushBack(&ai.sceneList)
}

func (ai *AIUser) Beat() {
}

func (ai *AIUser) SceneListNode() *List {
	return &ai.sceneList
}

func (ai *AIUser) Status() *ShipStatus {
	st := &ShipStatus {
		X: float64(ai.x),
		Y: float64(ai.y),
		Angle: float64(ai.angle),
		Rotate: ai.rotate,
		Move:ai.move,
		Hp: float64(ai.hp),
		Id: ai.id,
	}
	return st
}

func (ai *AIUser) updateBeam(delta float64) {
	var tmp *list.Element
	var beam *Beam

	for b := ai.beamList.Front(); b != nil; {
		beam = b.Value.(*Beam)
		tmp = b.Next()

		if !beam.Update(delta) {
			ai.beamList.Remove(b)
			ai.beamMap = ai.beamMap &^ (1 << uint(beam.id))
		}

		b = tmp
	}
}

func (ai *AIUser) updatePosition(delta float64) {
	if ai.rotate == 2 {
		angle := ai.angle + int(80 * (delta / 1000));
		if angle >= 360 {
			angle = angle - 360;
		}
		ai.angle = angle
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
	ai.algo.Update(ai)

	ai.updateAction(delta)
	ai.updateBeam(delta)
	ai.updatePosition(delta)
}

func (ai *AIUser) CheckHit(target Player) bool {
	if ai.id == target.UserId() {
		return false
	}

	var beam *Beam
	for b := ai.beamList.Front(); b != nil; b = b.Next() {
		beam = b.Value.(*Beam)
		if !beam.Hit(target.Position()) {
			continue
		}

		fmt.Printf("%d hit target %d\n", ai.id, target.UserId())

		ai.beamList.Remove(b)
		ai.beamMap = ai.beamMap &^ (1 << uint(beam.id))
		ai.scene.broadStopBeam(ai, int(beam.id), 1)

		// hit
		return true
	}

	return false
}

func NewAIUser() *AIUser {
	ai := &AIUser{
		x: 480,
		y: 320,
		name: "AI",
		hp: 100,
	}
	ai.beamList = list.New()
	InitList(&ai.sceneList, ai)
	ai.algo = NewAISimapleAlgo()
	return ai
}
