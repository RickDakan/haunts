package game

import (
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/base"
  "path/filepath"
)

type UiStart struct {
  layout UiStartLayout

  region gui.Region

  buttons []*Button

  // Position of the mouse
  mx,my int

  game_panel *GamePanel
}

type UiStartLayout struct {
  History, Random, Challenge, Replay Button
}

func makeChallegeSelection(_gp interface{}) {
  gp := _gp.(*GamePanel)
  gp.AnchorBox = gui.MakeAnchorBox(gui.Dims{1024,700})
  select_map, err := MakeUiSelectMap(gp)
  if err != nil {
    base.Error().Printf("Unable to make Map Selector: %v", err)
    return
  }
  gp.AnchorBox.AddChild(select_map, gui.Anchor{0, 0, 0, 0})
}

func MakeUiStart(gp *GamePanel) (gui.Widget, error) {
  var ui UiStart

  datadir := base.GetDataDir()
  err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "start.json"), "json", &ui.layout)
  if err != nil {
    return nil, err
  }

  ui.game_panel = gp

  ui.buttons = []*Button{
    &ui.layout.Challenge,
    &ui.layout.History,
    &ui.layout.Replay,
    &ui.layout.Random,
  }

  ui.layout.Challenge.f = makeChallegeSelection

  ui.layout.History.f = func(interface {}) {
    base.Log().Printf("History!")
  }
  ui.layout.Replay.f = func(interface {}) {
    base.Log().Printf("Replay!")
  }
  ui.layout.Random.f = func(interface {}) {
    base.Log().Printf("Random!")
  }

  ui.region.Dx = 1024
  ui.region.Dy = 768

  return &ui, nil
}

func (ui *UiStart) String() string {
  return "ui start"
}

func (ui *UiStart) Requested() gui.Dims {
  return gui.Dims{1024, 768}
}

func (ui *UiStart) Expandable() (bool, bool) {
  return false, false
}

func (ui *UiStart) Rendered() gui.Region {
  return ui.region
}

func (ui *UiStart) Think(g *gui.Gui, dt int64) {
  ui.mx, ui.my = gin.In().GetCursor("Mouse").Point()
}

func (ui *UiStart) Respond(g *gui.Gui, group gui.EventGroup) bool {
  if found, event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    for _, button := range ui.buttons {
      if button.handleClick(ui.mx, ui.my, ui.game_panel) {
        return true
      }
    }
  }
  for _, button := range ui.buttons {
    if button.Respond(group, ui) {
      return true
    }
  }
  return false
}

func (ui *UiStart) Draw(region gui.Region) {
  for _, button := range ui.buttons {
    button.RenderAt(region.X, region.Y, ui.mx, ui.my)
  }
}

func (ui *UiStart) DrawFocused(region gui.Region) {
  
}
