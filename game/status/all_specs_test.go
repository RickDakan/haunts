package conditions_test

import (
  "gospec"
  "testing"
)

func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(ConditionsSpec)
  gospec.MainGoTest(r, t)
}
