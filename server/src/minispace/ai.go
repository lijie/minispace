package minispace

type AIUser struct {
	id int
	name string
	sceneList List
}

func (ai *AIUser) SendClient(msg *Msg) error {
	return nil
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

func (ai *AIUser) SceneListNode() *List {
	return &ai.sceneList
}

func (ai *AIUser) Status() *ShipStatus {
	return nil
}

func (ai *AIUser) Update(delta float64) {
}

func (ai *AIUser) CheckHitAll(activeList *List) {
}
