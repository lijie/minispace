package minispace

import "math/rand"
import "time"

type aiSimpleAlgo struct {
	shootDelta float64
	rotateDelta float64
	rander *rand.Rand
}

func (algo *aiSimpleAlgo) Update(ai AIAction, dt float64) {
	// always keep moving
	ai.ActMove(1)

	if algo.rotateDelta >= 1000 {
		var value uint64
		value = uint64(algo.rander.Uint32()) * uint64(300) / uint64(^uint32(0))

		if value <= 100 {
			// just do right rotate
			ai.ActRotate(2)
		} else if value <= 200 {
			ai.ActRotate(1)
		} else {
			ai.ActRotate(0)
		}
		algo.rotateDelta = 0
	} else {
		algo.rotateDelta += dt
	}

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

var aiSimpleModVar = 1
func NewAISimapleAlgo() *aiSimpleAlgo {
	algo := &aiSimpleAlgo{
		rander: rand.New(rand.NewSource(time.Now().Unix() % int64(aiSimpleModVar))),
	}
	aiSimpleModVar += 73
	return algo
}
