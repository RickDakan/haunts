package game

import (
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/haunts/game/hui"
  "path/filepath"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/util/algorithm"
  "github.com/runningwild/opengl/gl"
)

type iconWithText struct {
  Name string
  Icon texture.Object
}
func (c *iconWithText) Draw(hovered, selected, selectable bool, region gui.Region) {
  var f float64
  switch {
  case selected:
    f = 1.0
  case hovered || selectable:
    f = 0.6
  default:
    f = 0.4
  }
  gl.Color4d(f, f, f, 1)
  c.Icon.Data().RenderNatural(region.X, region.Y)
  if hovered && selectable {
    if selected {
      gl.Color4d(1, 0, 0, 0.3)
    } else {
      gl.Color4d(1, 0, 0, 1)
    }
    gl.Disable(gl.TEXTURE_2D)
    gl.Begin(gl.LINES)
    x := region.X + 1
    y := region.Y + 1
    x2 := region.X + region.Dx - 1
    y2 := region.Y + region.Dy - 1

    gl.Vertex2i(x, y)
    gl.Vertex2i(x, y2)

    gl.Vertex2i(x, y2)
    gl.Vertex2i(x2, y2)

    gl.Vertex2i(x2, y2)
    gl.Vertex2i(x2, y)

    gl.Vertex2i(x2, y)
    gl.Vertex2i(x, y)
    gl.End()
  }
}
func (c *iconWithText) Think(dt int64) {
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
    x,y := gin.In().GetCursor("Mouse").Point()
    p := gui.Point{x,y}
    if p.Inside(hb.Render_region) {
      hb.f()
      return true
    }
    return false
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

  roster_chooser *hui.RosterChooser

  layout explorerSetupLayout
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
    return a.(*Entity).Side() == SideExplorers
  }).([]*Entity)

  es.roster_chooser = hui.MakeRosterChooser([]hui.Option{
    &iconWithText{
      Name: "Thunder",
      Icon: texture.Object{ Path: base.Path(filepath.Join(base.GetDataDir(), "ui", "haunts_setup", "ghosts.png"))},
    },
    &iconWithText{
      Name: "Monkey",
      Icon: texture.Object{ Path: base.Path(filepath.Join(base.GetDataDir(), "ui", "haunts_setup", "ghosts2.png"))},
    },
    &iconWithText{
      Name: "Chalice",
      Icon: texture.Object{ Path: base.Path(filepath.Join(base.GetDataDir(), "ui", "haunts_setup", "ghosts.png"))},
    },
    &iconWithText{
      Name: "Grandiose",
      Icon: texture.Object{ Path: base.Path(filepath.Join(base.GetDataDir(), "ui", "cute1.png"))},
    },
    &iconWithText{
      Name: "Lightning",
      Icon: texture.Object{ Path: base.Path(filepath.Join(base.GetDataDir(), "ui", "cute2.png"))},
    },
    &iconWithText{
      Name: "Orangutan",
      Icon: texture.Object{ Path: base.Path(filepath.Join(base.GetDataDir(), "ui", "cute3.png"))},
    },
    &iconWithText{
      Name: "Mead",
      Icon: texture.Object{ Path: base.Path(filepath.Join(base.GetDataDir(), "ui", "cute4.png"))},
    },
    &iconWithText{
      Name: "Grandeure",
      Icon: texture.Object{ Path: base.Path(filepath.Join(base.GetDataDir(), "ui", "cute5.png"))},
    },
  },
  hui.SelectAtMostN(3),
  func(m map[int]bool) {
    base.Log().Printf("complete: %v", m)
    es.AnchorBox.RemoveChild(es.roster_chooser)
  },
  )

  es.AnchorBox = gui.MakeAnchorBox(gui.Dims{1024, 768})
  es.purpose_table = gui.MakeVerticalTable()
  for i := range es.layout.Purposes {
    purpose := es.layout.Purposes[i]
    f := func() {
      base.Log().Printf("Roster: %v", purpose)
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

func (es *explorerSetup) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if es.AnchorBox.Respond(ui, group) {
    return true
  }
  return false
}

func (es *explorerSetup) Draw(r gui.Region) {
  es.BasicZone.Render_region = r
  es.AnchorBox.Draw(r)
}

func (es *explorerSetup) String() string {
  return "explorer setup"
}

