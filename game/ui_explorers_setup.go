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
  "github.com/runningwild/mathgl"
)

type entityLabel struct {
  ent *Entity
  hovered, selected, selectable bool
}
func makeEntLabel(ent *Entity) *entityLabel {
  var e entityLabel
  e.ent = ent
  e.ent.TurnToFace(1, -2)
  return &e
}
func (e *entityLabel) Draw(hovered, selected, selectable bool, region gui.Region) {
  e.hovered = hovered
  e.selected = selected
  e.selectable = selectable
  gl.Disable(gl.TEXTURE_2D)
  var f float64
  switch {
  case selected:
    f = 1
  case hovered && selectable:
    f = 0.8
  case selectable:
    f = 0.5
  default:
    f = 0.3
  }
  gl.Color4d(f, f, f, 1)
  gl.Begin(gl.QUADS)
    gl.Vertex2i(region.X, region.Y)
    gl.Vertex2i(region.X, region.Y + region.Dy)

    gl.Vertex2i(region.X, region.Y + region.Dy)
    gl.Vertex2i(region.X + region.Dx, region.Y + region.Dy)

    gl.Vertex2i(region.X + region.Dx, region.Y + region.Dy)
    gl.Vertex2i(region.X + region.Dx, region.Y)

    gl.Vertex2i(region.X + region.Dx, region.Y)
    gl.Vertex2i(region.X, region.Y)
  gl.End()
  d := base.GetDictionary(15)
  gl.Color4d(0, 0, 0, 1)
  d.RenderString(e.ent.Name, float64(region.X) + 210, float64(region.Y) + 100 - d.MaxHeight()/2, 0, d.MaxHeight(), gui.Center)
  if selectable || selected {
    f = 1
  }
  gl.Color4d(f, f, f, 1)
  e.ent.Render(mathgl.Vec2{float32(region.X + 55), float32(region.Y) + 20}, 100)
}
func (e *entityLabel) Think(dt int64) {
  if e.hovered && e.selectable {
    e.ent.sprite.Sprite().Command("move")
  } else {
    e.ent.sprite.Sprite().Command("stop")
  }
  if e.selected || e.selectable {
    e.ent.Think(dt)
  }
}

type iconWithText struct {
  Name string
  Icon texture.Object
  Data interface{}
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

  purpose_table gui.Widget

  roster_chooser *hui.RosterChooser
  gear_chooser   gui.Widget

  ents []*Entity

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

  var roster []hui.Option
  for _, ent := range ents {
    roster = append(roster, makeEntLabel(ent))
  }

  es.roster_chooser = hui.MakeRosterChooser(roster,
  hui.SelectInRange(3,3),
  func(m map[int]bool) {
    es.ents = es.ents[0:0]
    for i := range m {
      es.ents = append(es.ents, roster[i].(*entityLabel).ent)
    }
    es.AnchorBox.RemoveChild(es.roster_chooser)
    es.gear_chooser = es.makeGearChooser(game, 0)
    if es.gear_chooser == nil {
      es.startGame(game)
    } else {
      es.AnchorBox.AddChild(es.gear_chooser, gui.Anchor{0.5, 0.5, 0.5, 0.5})
    }
  },
  nil,
  )

  es.AnchorBox = gui.MakeAnchorBox(gui.Dims{1024, 768})

  purposes := []hui.Option{
    &iconWithText{
      Name: "Relic",
      Icon: texture.Object{ Path: base.Path(filepath.Join(base.GetDataDir(), "ui", "explorer_setup", "relic.png"))},
      Data: PurposeRelic,
    },
    &iconWithText{
      Name: "Cleanse",
      Icon: texture.Object{ Path: base.Path(filepath.Join(base.GetDataDir(), "ui", "explorer_setup", "cleanse.png"))},
      Data: PurposeCleanse,
    },
    &iconWithText{
      Name: "Mystery",
      Icon: texture.Object{ Path: base.Path(filepath.Join(base.GetDataDir(), "ui", "explorer_setup", "mystery.png"))},
      Data: PurposeMystery,
    },
  }
  es.purpose_table = hui.MakeRosterChooser(purposes, hui.SelectExactlyOne, func(m map[int]bool) {
    for k := range m {
      game.Purpose = purposes[k].(*iconWithText).Data.(Purpose)
      break
    }
    base.Log().Printf("Selected %d", game.Purpose)
    es.AnchorBox.RemoveChild(es.purpose_table)
    es.AnchorBox.AddChild(es.roster_chooser, gui.Anchor{0.5, 0.5, 0.5, 0.5})
  },
  nil)
  es.AnchorBox.AddChild(es.purpose_table, gui.Anchor{0.5, 0.5, 0.5, 0.5})

  return &es, nil
}

type gearChooser struct {
  *gui.AnchorBox

  ent     *Entity
  chooser *hui.RosterChooser
}
func makeGearChooser(ent *Entity, chooser *hui.RosterChooser) *gearChooser {
  var gc gearChooser
  gc.AnchorBox = gui.MakeAnchorBox(gui.Dims{1024, 768})
  gc.ent = ent
  gc.chooser = chooser
  gc.AnchorBox.AddChild(gc.chooser, gui.Anchor{0.5,0.5,0.5,0.5})
  return &gc
}
func (g *gearChooser) Draw(region gui.Region) {
  g.AnchorBox.Draw(region)
  x := region.X + g.chooser.Render_region.X / 2 - 50
  y := region.Y + region.Dy / 2
  gl.Color4ub(255, 255, 255, 255)
  g.ent.Render(mathgl.Vec2{float32(x), float32(y)}, 100)
  d := base.GetDictionary(15)
  gl.Color4ub(255, 255, 255, 255)
  d.RenderString(g.ent.Name, float64(x + 50), float64(y) - d.MaxHeight(), 0, d.MaxHeight(), gui.Center)
}

func (es *explorerSetup) startGame(game *Game) {
  game.PlaceInitialExplorers(es.ents)
  game.OnRound()
  for i := range game.Ents {
    // Something might still be walking, so lets just stop everything before
    // we move on.
    game.Ents[i].sprite.sp.Command("stop")
  }
}

func (es *explorerSetup) makeGearChooser(game *Game, explorer_index int) gui.Widget {
  if explorer_index > 2 {
    return nil
  }
  ent := es.ents[explorer_index]
  if len(ent.ExplorerEnt.Gear_names) == 0 {
    return es.makeGearChooser(game, explorer_index + 1)
  }
  var gear []hui.Option
  for _, name := range ent.ExplorerEnt.Gear_names {
    g := MakeGear(name)
    gear = append(gear, &iconWithText{
      Name: g.Name,
      Icon: g.Large_icon,
      Data: g,
    })
  }
  rc := hui.MakeRosterChooser(gear, hui.SelectExactlyOne, func(m map[int]bool) {
    for index := range m {
      ent.ExplorerEnt.Gear = MakeGear(ent.ExplorerEnt.Gear_names[index])
    }
    if explorer_index < 2 {
      es.AnchorBox.RemoveChild(es.gear_chooser)
      es.gear_chooser = es.makeGearChooser(game, explorer_index + 1)
      es.AnchorBox.AddChild(es.gear_chooser, gui.Anchor{0.5, 0.5, 0.5, 0.5})
    } else {
      es.startGame(game)
    }
  },
  nil,
  )
  return makeGearChooser(ent, rc)
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

