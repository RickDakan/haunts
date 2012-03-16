package game

import (
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
  "path/filepath"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/util/algorithm"
  "github.com/runningwild/mathgl"
  "github.com/runningwild/opengl/gl"
)

type iconWithText struct {
  Name string
  Icon texture.Object
}

type rosterLayout struct {
  Dx, Dy int
  Buffer int
  Ent struct {
    X, Y int
  }
}

type explorerSetupLayout struct {
  Purposes []iconWithText
  Purpose struct {
    Dx, Dy int
  }
  Roster rosterLayout
}

type hoverButton struct {
  gui.BasicZone
  text string
  icon texture.Object
  over bool
  amt  float64
  f    func()
}
func makeHoverButton(dx,dy int, text string, icon texture.Object, f func()) *hoverButton {
  var hb hoverButton
  hb.Request_dims.Dx = dx
  hb.Request_dims.Dy = dy
  hb.text = text
  hb.icon = icon
  hb.f = f
  return &hb
}
func (hb *hoverButton) String() string {
  return "hover button"
}
func (hb *hoverButton) Think(ui *gui.Gui, t int64) {
  x, y := gin.In().GetCursor("Mouse").Point()
  m := gui.Point{x, y}
  hb.over = m.Inside(hb.Render_region)
  frac := 0.9
  if hb.over {
    hb.amt = frac * hb.amt + (1-frac) * 1
  } else {
    hb.amt = frac * hb.amt + (1-frac) * 0.5
  }
}
func (hb *hoverButton) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if found, event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    hb.f()
  }
  return false
}
func (hb *hoverButton) Draw(r gui.Region) {
  hb.Render_region = r
  gl.Disable(gl.TEXTURE_2D)
  gl.Color4d(hb.amt, hb.amt, hb.amt, hb.amt)
  gl.Begin(gl.QUADS)
    gl.Vertex2i(r.X, r.Y)
    gl.Vertex2i(r.X, r.Y + r.Dy)
    gl.Vertex2i(r.X + r.Dx, r.Y + r.Dy)
    gl.Vertex2i(r.X + r.Dx, r.Y)
  gl.End()
  hb.icon.Data().Bind()
  gl.Enable(gl.TEXTURE_2D)
  gl.Begin(gl.QUADS)
    gl.TexCoord2d(0, 0)
    gl.Vertex2i(r.X, r.Y)

    gl.TexCoord2d(0, -1)
    gl.Vertex2i(r.X, r.Y + r.Dy)

    gl.TexCoord2d(1, -1)
    gl.Vertex2i(r.X + r.Dx/2, r.Y + r.Dy)

    gl.TexCoord2d(1, 0)
    gl.Vertex2i(r.X + r.Dx/2, r.Y)
  gl.End()
}
func (hb *hoverButton) DrawFocused(r gui.Region) {
}



// This is the UI that the explorers player uses to select his roster at the
// beginning of the game.  It will necessarily be centered on the screen
type explorerSetup struct {
  *gui.AnchorBox

  purpose_table *gui.VerticalTable

  roster_chooser *rosterChooser

  layout explorerSetupLayout
}

type rosterChooser struct {
  gui.BasicZone
  layout rosterLayout
  game *Game
  ents []*Entity

  // So we can give a dt to the ents to animate them
  last_think int64

  // What ent we are currently focused on
  focus int

  // As we move the focus around we gradually move our view to smoothly
  // adjust
  focus_pos float64

  // What ents we have currently selected
  selected map[int]bool
  selected_order [3]*Entity
}

func makeRosterChooser(layout rosterLayout, game *Game, ents []*Entity) *rosterChooser {
  var rc rosterChooser
  rc.BasicZone.Request_dims.Dx = 2*layout.Dx + layout.Buffer
  rc.BasicZone.Request_dims.Dy = 3*layout.Dy
  rc.layout = layout
  rc.game = game
  rc.ents = ents
  rc.selected = make(map[int]bool)
  return &rc
}

func MakeExplorerSetupBar(game *Game) (*explorerSetup, error) {
  var es explorerSetup
  datadir := base.GetDataDir()
  err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "explorer_setup", "config.json"), "json", &es.layout)
  if err != nil {
    return nil, err
  }

  names := base.GetAllNamesInRegistry("entities")
  ents := algorithm.Map(names, []*Entity{}, func(a interface{}) interface{} {
    return MakeEntity(a.(string), game)
  }).([]*Entity)
  ents = algorithm.Choose(ents, func(a interface{}) bool {
    return a.(*Entity).Side == game.Side
  }).([]*Entity)

  es.roster_chooser = makeRosterChooser(es.layout.Roster, game, ents)
  es.AnchorBox = gui.MakeAnchorBox(gui.Dims{1024, 768})
  es.purpose_table = gui.MakeVerticalTable()
  for i := range es.layout.Purposes {
    purpose := es.layout.Purposes[i]
    f := func() {
      es.AnchorBox.RemoveChild(es.purpose_table)
      es.AnchorBox.AddChild(es.roster_chooser, gui.Anchor{0.5, 0.5, 0.5, 0.5})
      switch es.layout.Purposes[i].Name {
      case "Relic":
        game.Purpose = PurposeRelic
      case "Cleanse":
        game.Purpose = PurposeCleanse
      case "Mystery":
        game.Purpose = PurposeMystery
      }
    }
    es.purpose_table.AddChild(makeHoverButton(es.layout.Purpose.Dx, es.layout.Purpose.Dy, purpose.Name, purpose.Icon, f))
  }
  es.AnchorBox.AddChild(es.purpose_table, gui.Anchor{0.5, 0.5, 0.5, 0.5})

  return &es, nil
}

func (es *explorerSetup) Think(ui *gui.Gui, t int64) {
  es.AnchorBox.Think(ui, t)
}

func (rc *rosterChooser) Think(ui *gui.Gui, t int64) {
  var dt int64
  if rc.last_think != 0 {
    dt = t - rc.last_think
  }
  rc.last_think = t
  for i := range rc.ents {
    rc.ents[i].Think(dt)
  }

  max := len(rc.ents)
  if rc.focus < 0 {
    rc.focus = 0
  }
  if rc.focus >= max {
    rc.focus = max - 1
  }
  frac := 0.8
  rc.focus_pos = frac * rc.focus_pos + (1-frac) * float64(rc.focus)
}

func (es *explorerSetup) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if es.AnchorBox.Respond(ui, group) {
    return true
  }
  return false
}

func (rc *rosterChooser) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if found, event := group.FindEvent('l'); found && event.Type == gin.Press {
    rc.focus+=3
    return true
  }
  if found, event := group.FindEvent('o'); found && event.Type == gin.Press {
    rc.focus-=3
    return true
  }
  if found, event := group.FindEvent(gin.Return); found && event.Type == gin.Press {
    if len(rc.selected) == len(rc.selected_order) {
      rc.game.PlaceInitialExplorers(rc.selected_order[:])
      rc.game.OnRound()
    }
    return true
  }

  cursor := gin.In().GetCursor("Mouse")
  if found, event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    x,y := cursor.Point()

    // Throw out the click if it is outside of this window
    if x < rc.Render_region.X { return false }
    if y < rc.Render_region.Y { return false }
    if x >= rc.Render_region.X + rc.Render_region.Dx { return false }
    if y >= rc.Render_region.Y + rc.Render_region.Dy { return false }

    // Throw out the click if it is in the buffer region
    if x >= rc.Render_region.X + rc.layout.Dx &&
       x < rc.Render_region.X + rc.layout.Dx + rc.layout.Buffer {
        return false
      }

    if x < rc.Render_region.X + rc.layout.Dx {
      // If it is on the left figure out what index was clicked and either add
      // or remove that ent if possible
      off := float64(rc.Render_region.Y + 2*rc.layout.Dy - y) / float64(rc.layout.Dy)
      off += rc.focus_pos
      index := int(off)
      if index >= 0 && index < len(rc.ents) {
        if rc.selected[index] {
          delete(rc.selected, index)
          for i := 0; i < len(rc.selected_order); i++ {
            if rc.selected_order[i] == rc.ents[index] {
              rc.selected_order[i] = nil
            }
          }
        } else {
          if len(rc.selected) < 3 {
            rc.selected[index] = true
            for i := 0; i < len(rc.selected_order); i++ {
              if rc.selected_order[i] == nil {
                rc.selected_order[i] = rc.ents[index]
                break
              }
            }
          }
        }
      }
    } else {
      // The click is in the other region, we can only deselect ents from here
      off := float64(rc.Render_region.Y + 3*rc.layout.Dy - y) / float64(rc.layout.Dy)
      index := int(off)
      if index >= 0 && index < len(rc.selected_order) && rc.selected_order[index] != nil {
        for i := range rc.ents {
          if rc.ents[i] == rc.selected_order[index] {
            rc.selected_order[index] = nil
            delete(rc.selected, i)
            break
          }
        }
      }
    }
  }
  return false
}

func (es *explorerSetup) Draw(r gui.Region) {
  es.BasicZone.Render_region = r
  es.AnchorBox.Draw(r)
}

func (rc *rosterChooser) Draw(r gui.Region) {
  rc.Render_region = r
  gl.Disable(gl.TEXTURE_2D)
  r.PushClipPlanes()
  defer r.PopClipPlanes()
  x := float64(r.X)
  y := float64(r.Y) + float64(rc.layout.Dy) * (2 + rc.focus_pos)
  for i := -1; i <= len(rc.ents); i++ {
    if rc.selected[i] {
      gl.Color4d(1, 1, 1, 1)
    } else {
      gl.Color4d(0.5, 0.5, 0.5, 1)
    }
    dx := float64(rc.layout.Dx)
    dy := float64(rc.layout.Dy)
    gl.Disable(gl.TEXTURE_2D)
    gl.Begin(gl.QUADS)
      gl.Vertex2d(x, y)
      gl.Vertex2d(x, y + dy)
      gl.Vertex2d(x + dx, y + dy)
      gl.Vertex2d(x + dx, y)
    gl.End()
    if i >= 0 && i < len(rc.ents) {
      rc.ents[i].Render(mathgl.Vec2{float32(r.X + rc.layout.Ent.X), float32(int(y) + rc.layout.Ent.Y)}, 100)
      d := base.GetDictionary(15)
      gl.Color4d(0, 0, 0, 1)
      d.RenderString(rc.ents[i].Name, x + float64(rc.layout.Ent.X)*2, y + float64(rc.layout.Dy)/2, 0, d.MaxHeight(), gui.Left)
    }
    y -= dy
  }

  x = float64(r.X + rc.layout.Dx + rc.layout.Buffer)
  y = float64(r.Y) + float64(rc.layout.Dy) * 3
  for i := range rc.selected_order {
    dx := float64(rc.layout.Dx)
    dy := float64(rc.layout.Dy)
    y -= dy
    if rc.selected_order[i] != nil {
      gl.Color4d(1, 1, 1, 1)
    } else {
      continue
    }
    gl.Disable(gl.TEXTURE_2D)
    gl.Begin(gl.QUADS)
      gl.Vertex2d(x, y)
      gl.Vertex2d(x, y + dy)
      gl.Vertex2d(x + dx, y + dy)
      gl.Vertex2d(x + dx, y)
    gl.End()
    if i >= 0 && i < len(rc.ents) {
      rc.selected_order[i].Render(mathgl.Vec2{float32(int(x) + rc.layout.Ent.X), float32(int(y) + rc.layout.Ent.Y)}, 100)
      d := base.GetDictionary(15)
      gl.Color4d(0, 0, 0, 1)
      d.RenderString(rc.selected_order[i].Name, x + float64(rc.layout.Ent.X)*2, y + float64(rc.layout.Dy)/2, 0, d.MaxHeight(), gui.Left)
    }
  }


  gl.Disable(gl.TEXTURE_2D)
  dx := rc.layout.Dx
  gl.Begin(gl.QUADS)
    gl.Color4d(0, 0, 0, 0)
    gl.Vertex2i(r.X, r.Y + rc.layout.Dy)
    gl.Vertex2i(r.X + dx, r.Y + rc.layout.Dy)
    gl.Color4d(0, 0, 0, 1)
    gl.Vertex2i(r.X + dx, r.Y)
    gl.Vertex2i(r.X, r.Y)

    gl.Color4d(0, 0, 0, 0)
    gl.Vertex2i(r.X + dx, r.Y + 2*rc.layout.Dy)
    gl.Vertex2i(r.X, r.Y + 2*rc.layout.Dy)
    gl.Color4d(0, 0, 0, 1)
    gl.Vertex2i(r.X, r.Y + 3*rc.layout.Dy)
    gl.Vertex2i(r.X + dx, r.Y + 3*rc.layout.Dy)
  gl.End()
}

func (rc *rosterChooser) DrawFocused(gui.Region) {

}

func (rc *rosterChooser) String() string {
  return "roster chooser"
}


func (es *explorerSetup) String() string {
  return "explorer setup"
}

