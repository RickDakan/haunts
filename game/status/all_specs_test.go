package conditions_test

import (
  "gospec"
  "testing"
  "haunts/game/status"
)

func TestAllSpecs(t *testing.T) {
  status.RegisterAllConditions()
  r := gospec.NewRunner()
  r.AddSpec(ConditionsSpec)
  gospec.MainGoTest(r, t)
}
