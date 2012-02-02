package action_test

import (
  "gospec"
  "testing"
)

func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(ActionSpec)
  gospec.MainGoTest(r, t)
}
