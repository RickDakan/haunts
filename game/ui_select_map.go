package game

import (
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/house"
  "github.com/runningwild/haunts/game/hui"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
  gl "github.com/chsc/gogl/gl21"
  "path/filepath"
)

type UiSelectMap struct {
  layout UiSelectMapLayout

  region gui.Region

  chooser *hui.RosterChooser
}

type MapOption struct {
  layout *UiSelectMapLayout

  house_def *house.HouseDef
}
func (mo *MapOption) Draw(hovered, selected, selectable bool, region gui.Region) {
  var s byte
  switch {
  case selected:
    s = 255
  case hovered && selectable:
    s = 205
  case selectable:
    s = 127
  default:
    s = 75
  }
  gl.Color4ub(s, s, s, 255)
  icon := mo.house_def.Icon.Data()
  if icon.Dx() == 0 {
    icon = mo.layout.Default_icon.Data()
  }
  icon.RenderNatural(region.X, region.Y)
  gl.Color4ub(0, 0, 0, 255)
  d := base.GetDictionary(15)
  d.RenderString(mo.house_def.Name, float64(region.X), float64(region.Y), 0, d.MaxHeight(), gui.Left)
}
func (mo *MapOption) Think(dt int64) {
}

type UiSelectMapLayout struct {
  Default_icon texture.Object
}

func MakeUiSelectMap(gp *GamePanel) (gui.Widget, error) {
  var ui UiSelectMap

  datadir := base.GetDataDir()
  err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "select_map", "config.json"), "json", &ui.layout)
  if err != nil {
    return nil, err
  }

  ui.region.Dx = 1024
  ui.region.Dy = 768
  var options []hui.Option
  for _, name := range base.GetAllNamesInRegistry("houses") {
    var mo MapOption
    base.Log().Printf("Making house: %s", name)
    mo.house_def = house.MakeHouseFromName(name)
    mo.layout = &ui.layout
    options = append(options, &mo)
  }
  chooser := hui.MakeRosterChooser(options, hui.SelectExactlyOne, func(m map[int]bool) {
    var index int
    for index = range m {
      break
    }
    gp.AnchorBox = gui.MakeAnchorBox(gui.Dims{1024,700})
    select_side, err := MakeUiSelectSide(gp, options[index].(*MapOption).house_def)
    if err != nil {
      base.Error().Printf("Unable to make Side Selector: %v", err)
      return
    }
    gp.AnchorBox.AddChild(select_side, gui.Anchor{0, 0, 0, 0})
  })
  ui.chooser = chooser

  return &ui, nil
}

func (ui *UiSelectMap) String() string {
  return "ui start"
}

func (ui *UiSelectMap) Requested() gui.Dims {
  return gui.Dims{1024, 768}
}

func (ui *UiSelectMap) Expandable() (bool, bool) {
  return false, false
}

func (ui *UiSelectMap) Rendered() gui.Region {
  return ui.region
}

func (ui *UiSelectMap) Think(g *gui.Gui, dt int64) {
  ui.chooser.Think(g, dt)
}

func (ui *UiSelectMap) Respond(g *gui.Gui, group gui.EventGroup) bool {
  return ui.chooser.Respond(g, group)
}

func (ui *UiSelectMap) Draw(region gui.Region) {
  ui.chooser.Draw(region)
}

func (ui *UiSelectMap) DrawFocused(region gui.Region) {
  ui.chooser.Draw(region)
}
