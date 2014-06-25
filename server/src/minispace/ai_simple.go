package minispace

type aiSimpleAlgo struct {
	shootDelta float64
}

func (algo *aiSimpleAlgo) Update(ai AIAction, dt float64) {
	// just do right rotate
	ai.ActRotate(2)
	if algo.shootDelta >= 1000 {
		// shoot if possible
		ai.ActShoot()
		algo.shootDelta = 0
	} else {
		algo.shootDelta += dt
	}
}

func (algo *aiSimpleAlgo) Name() string {
	return "ai_simple_algo"
}

func NewAISimapleAlgo() *aiSimpleAlgo {
	return &aiSimpleAlgo{}
}
