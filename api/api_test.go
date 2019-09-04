package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	path := runTemplate(startScenario, &params{ScenarioID: 1, Speed: 2, MaxDelay: 3})
	assert.Equal(t, "/v1/replay/scenario/play/1?speed=2&max_delay=3&use_replay_timestamp=false", path)
}
