package game

import (
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/haunts/base"
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
    Text     struct {
      X, Y    int
      Size    int
      Spacing int
    }
  }
  Background texture.Object
}

type StartMenu struct {
  sub_menu gui.Widget
  layout   startLayout
  region   gui.Region
  buttons  []*Button
  mx, my   int
}

func MakeStartMenu() (*StartMenu, error) {
  var sm StartMenu
  datadir := base.GetDataDir()
  err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "start", "layout.json"), "json", &sm.layout)
  if err != nil {
    return nil, err
  }
  sm.buttons = []*Button{
    &sm.layout.Menu.Continue,
    &sm.layout.Menu.Versus,
    &sm.layout.Menu.Story,
    &sm.layout.Menu.Settings,
  }
  sm.layout.Menu.Continue.f = func(interface{}) {}
  sm.layout.Menu.Versus.f = func(interface{}) {}
  sm.layout.Menu.Settings.f = func(interface{}) {}
  sm.layout.Menu.Story.f = func(interface{}) {
    var err error
    sm.sub_menu, err = MakeStoryMenu(&sm)
    if err != nil {
      base.Error().Printf("Unable to make Story Menu: %v", err)
      return
    }
    base.Log().Printf("Submenu: %v", sm.sub_menu)
  }
  return &sm, nil
}

func (sm *StartMenu) Requested() gui.Dims {
  if sm.sub_menu != nil {
    return sm.Requested()
  }
  return gui.Dims{1024, 768}
}

func (sm *StartMenu) Expandable() (bool, bool) {
  if sm.sub_menu != nil {
    return sm.sub_menu.Expandable()
  }
  return false, false
}

func (sm *StartMenu) Rendered() gui.Region {
  if sm.sub_menu != nil {
    return sm.sub_menu.Rendered()
  }
  return sm.region
}

func (sm *StartMenu) Think(g *gui.Gui, t int64) {
  if sm.sub_menu != nil {
    sm.sub_menu.Think(g, t)
    return
  }
  if sm.mx == 0 && sm.my == 0 {
    sm.mx, sm.my = gin.In().GetCursor("Mouse").Point()
  }
  for _, button := range sm.buttons {
    button.Think(sm.region.X, sm.region.Y, sm.mx, sm.my, t)
  }
}

func (sm *StartMenu) Respond(g *gui.Gui, group gui.EventGroup) bool {
  if sm.sub_menu != nil {
    return sm.sub_menu.Respond(g, group)
  }

  cursor := group.Events[0].Key.Cursor()
  if cursor != nil {
    sm.mx, sm.my = cursor.Point()
  }

  if found, event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    for _, button := range sm.buttons {
      if button.handleClick(sm.mx, sm.my, nil) {
        return true
      }
    }
  }
  return false
}

func (sm *StartMenu) Draw(region gui.Region) {
  if sm.sub_menu != nil {
    sm.sub_menu.Draw(region)
    return
  }
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
