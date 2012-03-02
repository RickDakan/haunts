package game

import (
  "fmt"
  "path/filepath"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/opengl/gl"
)

type Button struct {
  X,Y int
  Texture texture.Object

  // Color - brighter when the mouse is over it
  shade float64
}

func (b *Button) RenderAt(x,y,mx,my int) {
  b.Texture.Data().Bind()
  tdx :=  + b.Texture.Data().Dx
  tdy :=  + b.Texture.Data().Dy
  if mx >= x + b.X && mx < x + b.X + tdx && my >= y + b.Y && my < y + b.Y + tdy {
    b.shade = b.shade * 0.9 + 0.1
  } else {
    b.shade = b.shade * 0.9 + 0.04
  }
  gl.Color4d(1, 1, 1, b.shade)
  gl.Begin(gl.QUADS)
    gl.TexCoord2d(0, 0)
    gl.Vertex2i(x + b.X, y + b.Y)

    gl.TexCoord2d(0, -1)
    gl.Vertex2i(x + b.X, y + b.Y + tdy)

    gl.TexCoord2d(1,-1)
    gl.Vertex2i(x + b.X + tdx, y + b.Y + tdy)

    gl.TexCoord2d(1, 0)
    gl.Vertex2i(x + b.X + tdx, y + b.Y)
  gl.End()
}

type Center struct {
  X,Y int
}

type TextArea struct {
  X,Y           int
  Size          int
  Justification string
}

func (t *TextArea) RenderString(s string) {
  var just gui.Justification
  switch t.Justification {
  case "center":
    just = gui.Center
  case "left":
    just = gui.Left
  case "right":
    just = gui.Right
  default:
    base.Warn().Printf("Unknown justification '%s' in main gui bar.", t.Justification)
    t.Justification = "center"
  }
  px := float64(t.X)
  py := float64(t.Y)
  d := base.GetDictionary(t.Size)
  d.RenderString(s, px, py, 0, d.MaxHeight(), just)
}

type MainBarLayout struct {
  EndTurn     Button
  UnitLeft    Button
  UnitRight   Button
  ActionLeft  Button
  ActionRight Button

  CenterStillFrame Center

  Background texture.Object
  Divider    texture.Object
  Name   TextArea
  Ap     TextArea
  Hp     TextArea
  Corpus TextArea
  Ego    TextArea

  Conditions struct {
    X,Y,Height,Width,Size,Spacing float64
  }

  Actions struct {
    X,Y,Width,Icon_size float64
    Count int
  }
}

type MainBar struct {
  layout MainBarLayout
  region gui.Region

  Ent *Entity

  // Position of the mouse
  mx,my int
}

func MakeMainBar() (*MainBar, error) {
  var mb MainBar
  datadir := base.GetDataDir()
  err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "main_bar.json"), "json", &mb.layout)
  if err != nil {
    return nil, err
  }
  return &mb, nil
}
func (m *MainBar) Requested() gui.Dims {
  return gui.Dims{
    Dx: m.layout.Background.Data().Dx,
    Dy: m.layout.Background.Data().Dy,
  }
}

func (m *MainBar) Expandable() (bool, bool) {
  return false, false
}

func (m *MainBar) Rendered() gui.Region {
  return m.region
}


func (m *MainBar) Think(g *gui.Gui, t int64) {

}

func (m *MainBar) Respond(g *gui.Gui, group gui.EventGroup) bool {
  cursor := group.Events[0].Key.Cursor()
  if cursor != nil {
    m.mx, m.my = cursor.Point()
    x := m.region.X
    x2 := m.region.X + m.region.Dx
    y := m.region.Y
    y2 := m.region.Y + m.region.Dy
    if m.mx >= x && m.mx < x2 && m.my >= y && m.my < y2 {
      return true
    }
  }
  return false
}

func (m *MainBar) Draw(region gui.Region) {
  m.region = region
  gl.Enable(gl.TEXTURE_2D)
  m.layout.Background.Data().Bind()
  gl.Color4d(1, 1, 1, 1)
  gl.Begin(gl.QUADS)
    gl.TexCoord2d(0, 0)
    gl.Vertex2i(region.X, region.Y)

    gl.TexCoord2d(0, -1)
    gl.Vertex2i(region.X, region.Y + region.Dy)

    gl.TexCoord2d(1,-1)
    gl.Vertex2i(region.X + region.Dx, region.Y + region.Dy)

    gl.TexCoord2d(1, 0)
    gl.Vertex2i(region.X + region.Dx, region.Y)
  gl.End()

  m.layout.UnitLeft.RenderAt(region.X, region.Y, m.mx, m.my)
  m.layout.UnitRight.RenderAt(region.X, region.Y, m.mx, m.my)
  m.layout.ActionLeft.RenderAt(region.X, region.Y, m.mx, m.my)
  m.layout.ActionRight.RenderAt(region.X, region.Y, m.mx, m.my)

  if m.Ent != nil {
    gl.Color4d(1, 1, 1, 1)
    m.Ent.Still.Data().Bind()
    tdx := m.Ent.Still.Data().Dx
    tdy := m.Ent.Still.Data().Dy
    cx := region.X + m.layout.CenterStillFrame.X
    cy := region.Y + m.layout.CenterStillFrame.Y
    gl.Begin(gl.QUADS)
      gl.TexCoord2d(0, 0)
      gl.Vertex2i(cx - tdx / 2, cy - tdy / 2)

      gl.TexCoord2d(0, -1)
      gl.Vertex2i(cx - tdx / 2, cy + tdy / 2)

      gl.TexCoord2d(1,-1)
      gl.Vertex2i(cx + tdx / 2, cy + tdy / 2)

      gl.TexCoord2d(1, 0)
      gl.Vertex2i(cx + tdx / 2, cy - tdy / 2)
    gl.End()

    m.layout.Name.RenderString(m.Ent.Name)
    m.layout.Ap.RenderString(fmt.Sprintf("Ap:%d", m.Ent.Stats.ApCur()))
    m.layout.Hp.RenderString(fmt.Sprintf("Hp:%d", m.Ent.Stats.HpCur()))
    m.layout.Corpus.RenderString(fmt.Sprintf("Corpus:%d", m.Ent.Stats.Corpus()))
    m.layout.Ego.RenderString(fmt.Sprintf("Ego:%d", m.Ent.Stats.Ego()))

    gl.Color4d(1, 1, 1, 1)
    m.layout.Divider.Data().Bind()
    tdx = m.layout.Divider.Data().Dx
    tdy = m.layout.Divider.Data().Dy
    cx = region.X + m.layout.Name.X
    cy = region.Y + m.layout.Name.Y - 5
    gl.Begin(gl.QUADS)
      gl.TexCoord2d(0, 0)
      gl.Vertex2i(cx - tdx / 2, cy - tdy / 2)

      gl.TexCoord2d(0, -1)
      gl.Vertex2i(cx - tdx / 2, cy + (tdy + 1) / 2)

      gl.TexCoord2d(1,-1)
      gl.Vertex2i(cx + (tdx + 1) / 2, cy + (tdy + 1) / 2)

      gl.TexCoord2d(1, 0)
      gl.Vertex2i(cx + (tdx + 1) / 2, cy - tdy / 2)
    gl.End()
  }

  {
    gl.Color4d(1, 1, 1, 1)
    c := m.layout.Conditions
    d := base.GetDictionary(int(c.Size))
    ypos := c.Height + c.Y
    var r gui.Region
    r.X = int(c.X)
    r.Y = int(c.Y)
    r.Dx = int(c.Width)
    r.Dy = int(c.Height)
    r.PushClipPlanes()

    for _,s := range []string{"Blinded!", "Terrified", "Vexed!", "Offended!", "Hypnotized!", "Conned!", "Scorhed!"} {
      d.RenderString(s, c.X + c.Width / 2, ypos - float64(d.MaxHeight()), 0, d.MaxHeight(), gui.Center)
      ypos -= float64(d.MaxHeight())
      ypos += 3
    }

    r.PopClipPlanes()
  }

  spacing := m.layout.Actions.Icon_size * float64(m.layout.Actions.Count)
  spacing = m.layout.Actions.Width - spacing
  spacing /= float64(m.layout.Actions.Count - 1)
  xpos := m.layout.Actions.X
  s := m.layout.Actions.Icon_size
  gl.Disable(gl.TEXTURE_2D)
  for i := 0; i < 6; i++ {
    gl.Color4d(1, 0, 0, 1)
    gl.Begin(gl.QUADS)
      gl.Vertex3d(xpos, m.layout.Actions.Y, 0)
      gl.Vertex3d(xpos, m.layout.Actions.Y+s, 0)
      gl.Vertex3d(xpos+s, m.layout.Actions.Y+s, 0)
      gl.Vertex3d(xpos+s, m.layout.Actions.Y, 0)
      xpos += spacing + m.layout.Actions.Icon_size
    gl.End()
  }
}

func (m *MainBar) DrawFocused(region gui.Region) {

}

func (m *MainBar) String() string {
  return "main bar"
}

