package minispace

import "math/rand"
import "time"

type aiSimpleAlgo struct {
	movedt float64
	rander *rand.Rand
}

func (algo *aiSimpleAlgo) Update(ai AIAction, dt float64) {
	if (algo.movedt <= 0) {
		x := uint64(algo.rander.Uint32()) * uint64(960) / uint64(^uint32(0))
		y := uint64(algo.rander.Uint32()) * uint64(540) / uint64(^uint32(0))
		ai.ActMove(float64(x), float64(y))
		algo.movedt = 5000
	} else {
		algo.movedt -= dt
	}
}

func (algo *aiSimpleAlgo) Name() string {
	return "ai_simple_algo"
}

var aiSimpleModVar = 1
func NewAISimapleAlgo() *aiSimpleAlgo {
	algo := &aiSimpleAlgo{
		rander: rand.New(rand.NewSource(time.Now().Unix() % int64(aiSimpleModVar))),
	}
	aiSimpleModVar += 73
	return algo
}
