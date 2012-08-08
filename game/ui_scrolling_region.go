package game

import (
  "github.com/runningwild/glop/gui"
)

// SUPER simple scrolling region
type ScrollingRegion struct {
  Region gui.Region
  Height int
  pos    float64
  target float64
}

func (sr *ScrollingRegion) Move(amt int) {
  sr.target += float64(amt)
}
func (sr *ScrollingRegion) Top() int {
  return int(sr.pos)
}
func (sr *ScrollingRegion) Think(dt int64) {
  if sr.target > float64(sr.Height-sr.Region.Dy) {
    sr.target = float64(sr.Height - sr.Region.Dy)
  }
  if sr.target < 0 {
    sr.target = 0
  }
  sr.pos = doApproach(sr.pos, sr.target, dt)
}
