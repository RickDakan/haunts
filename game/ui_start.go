package game

import (
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/opengl/gl"
  "path/filepath"
)

type startLayout struct {
  Menu struct {
    X, Y     int
    Texture  texture.Object
    Continue Button
    Versus   Button
    Story    Button
    Settings Button
  }
  Background texture.Object
}

type StartMenu struct {
  layout  startLayout
  region  gui.Region
  buttons []ButtonLike
  mx, my  int
  last_t  int64
}

func InsertStartMenu(ui gui.WidgetParent) error {
  var sm StartMenu
  datadir := base.GetDataDir()
  err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "start", "layout.json"), "json", &sm.layout)
  if err != nil {
    return err
  }
  sm.buttons = []ButtonLike{
    &sm.layout.Menu.Continue,
    &sm.layout.Menu.Versus,
    &sm.layout.Menu.Story,
    &sm.layout.Menu.Settings,
  }
  sm.layout.Menu.Continue.f = func(interface{}) {}
  sm.layout.Menu.Versus.f = func(interface{}) {
    ui.RemoveChild(&sm)
    ui.AddChild(MakeGamePanel("versus/basic.lua", nil, nil))
  }
  sm.layout.Menu.Settings.f = func(interface{}) {}
  sm.layout.Menu.Story.f = func(interface{}) {
    ui.RemoveChild(&sm)
    err := InsertStoryMenu(ui)
    if err != nil {
      base.Error().Printf("Unable to make Story Menu: %v", err)
      return
    }
  }
  ui.AddChild(&sm)
  return nil
}

func (sm *StartMenu) Requested() gui.Dims {
  return gui.Dims{1024, 768}
}

func (sm *StartMenu) Expandable() (bool, bool) {
  return false, false
}

func (sm *StartMenu) Rendered() gui.Region {
  return sm.region
}

func (sm *StartMenu) Think(g *gui.Gui, t int64) {
  if sm.last_t == 0 {
    sm.last_t = t
    return
  }
  dt := t - sm.last_t
  sm.last_t = t

  if sm.mx == 0 && sm.my == 0 {
    sm.mx, sm.my = gin.In().GetCursor("Mouse").Point()
  }
  for _, button := range sm.buttons {
    button.Think(sm.region.X, sm.region.Y, sm.mx, sm.my, dt)
  }
}

func (sm *StartMenu) Respond(g *gui.Gui, group gui.EventGroup) bool {
  cursor := group.Events[0].Key.Cursor()
  if cursor != nil {
    sm.mx, sm.my = cursor.Point()
  }

  if found, event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    hit := false
    for _, button := range sm.buttons {
      if button.handleClick(sm.mx, sm.my, nil) {
        hit = true
      }
    }
    if hit {
      return true
    }
  } else {
    hit := false
    for _, button := range sm.buttons {
      if button.Respond(group, nil) {
        hit = true
      }
    }
    if hit {
      return true
    }
  }
  return false
}

func (sm *StartMenu) Draw(region gui.Region) {
  sm.region = region
  gl.Color4ub(255, 255, 255, 255)
  sm.layout.Background.Data().RenderNatural(sm.region.X, sm.region.Y)
  sm.layout.Menu.Texture.Data().RenderNatural(sm.region.X+sm.layout.Menu.X, sm.region.Y+sm.layout.Menu.Y)
  for _, button := range sm.buttons {
    button.RenderAt(sm.region.X, sm.region.Y)
  }
}

func (sm *StartMenu) DrawFocused(region gui.Region) {
}

func (sm *StartMenu) String() string {
  return "start menu"
}
