package minispace

import "fmt"
import "container/list"

type AIUser struct {
	id int
	name string
	scene *Scene
	sceneList List
	x, y, angle, hp int
	beamMap int
	beamList *list.List
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
		Hp: float64(ai.hp),
		Id: ai.id,
	}
	return st
}

func (ai *AIUser) Update(delta float64) {
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
	}
	ai.beamList = list.New()
	InitList(&ai.sceneList, ai)
	return ai
}
