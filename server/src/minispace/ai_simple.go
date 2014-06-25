package minispace

type aiSimpleAlgo struct {
}

func (algo *aiSimpleAlgo) Update(ai *AIUser) {
	// just do right rotate
	ai.rotate = 2
}

func (algo *aiSimpleAlgo) Name() string {
	return "ai_simple_algo"
}

func NewAISimapleAlgo() *aiSimpleAlgo {
	return &aiSimpleAlgo{}
}
