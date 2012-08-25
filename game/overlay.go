package game

import (
  gl "github.com/chsc/gogl/gl21"
  "github.com/runningwild/glop/gui"
)

type Overlay struct {
  region gui.Region
  game   *Game
}

func MakeOverlay(g *Game) gui.Widget {
  return &Overlay{game: g}
}

func (o *Overlay) Requested() gui.Dims {
  return gui.Dims{1024, 768}
}
func (o *Overlay) Expandable() (bool, bool) {
  return false, false
}
func (o *Overlay) Rendered() gui.Region {
  return o.region
}
func (o *Overlay) Think(g *gui.Gui, t int64) {
}
func (o *Overlay) Respond(g *gui.Gui, group gui.EventGroup) bool {
  return false
}
func (o *Overlay) Draw(region gui.Region) {
  o.region = region
  switch o.game.Side {
  case SideHaunt:
    if o.game.los.denizens.mode == LosModeBlind {
      return
    }
  case SideExplorers:
    if o.game.los.intruders.mode == LosModeBlind {
      return
    }
  default:
    return
  }
  for _, way := range o.game.Waypoints {
    if way.Side != o.game.Side {
      continue
    }
    cx := float32(way.X)
    cy := float32(way.Y)
    r := float32(way.Radius)
    cx1, cy1 := o.game.viewer.BoardToWindow(cx-r, cy-r)
    cx2, cy2 := o.game.viewer.BoardToWindow(cx-r, cy+r)
    cx3, cy3 := o.game.viewer.BoardToWindow(cx+r, cy+r)
    cx4, cy4 := o.game.viewer.BoardToWindow(cx+r, cy-r)
    gl.Color4ub(200, 0, 0, 128)
    gl.Disable(gl.TEXTURE_2D)
    gl.Begin(gl.QUADS)
    gl.Vertex2i(int32(cx1), int32(cy1))
    gl.Vertex2i(int32(cx2), int32(cy2))
    gl.Vertex2i(int32(cx3), int32(cy3))
    gl.Vertex2i(int32(cx4), int32(cy4))
    gl.End()
  }
}
func (o *Overlay) DrawFocused(region gui.Region) {
  o.Draw(region)
}
func (o *Overlay) String() string {
  return "overlay"
}
