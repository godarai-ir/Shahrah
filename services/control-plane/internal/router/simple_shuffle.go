package router

import (
	"errors"
	"math/rand"
)

// SimpleShuffle implements RoutingStrategy using simple random selection.
type SimpleShuffle struct{}

// NewSimpleShuffle creates a new SimpleShuffle strategy.
func NewSimpleShuffle() *SimpleShuffle {
	return &SimpleShuffle{}
}

// Name returns the strategy name.
func (s *SimpleShuffle) Name() string {
	return "simple-shuffle"
}

// SelectDeployment selects a random deployment from healthy ones.
func (s *SimpleShuffle) SelectDeployment(deployments []*Deployment, model string) (*Deployment, error) {
	if len(deployments) == 0 {
		return nil, errors.New("no healthy deployments")
	}

	// Check if weights are configured for weighted random selection
	hasWeight := false

	for _, d := range deployments {
		if d.Weight > 0 {
			hasWeight = true
			break
		}
	}

	if hasWeight {
		return weightedRandomPick(deployments), nil
	}

	return deployments[rand.Intn(len(deployments))], nil
}
