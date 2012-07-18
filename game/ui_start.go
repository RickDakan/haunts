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
}

type StartMenu struct {
  layout  startLayout
  region  gui.Region
  back    *texture.Data
  menu    *texture.Data
  buttons []*Button
  mx, my  int
}

func MakeStartMenu() (*StartMenu, error) {
  var sm StartMenu
  datadir := base.GetDataDir()
  err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "start", "layout.json"), "json", &sm.layout)
  if err != nil {
    return nil, err
  }
  sm.back = texture.LoadFromPath(filepath.Join(datadir, "ui", "start", "start.png"))
  sm.menu = texture.LoadFromPath(filepath.Join(datadir, "ui", "start", "menu.png"))
  sm.buttons = []*Button{
    &sm.layout.Menu.Continue,
    &sm.layout.Menu.Versus,
    &sm.layout.Menu.Story,
    &sm.layout.Menu.Settings,
  }
  return &sm, nil
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
  if sm.mx == 0 && sm.my == 0 {
    sm.mx, sm.my = gin.In().GetCursor("Mouse").Point()
  }
  for _, button := range sm.buttons {
    button.Think(sm.region.X, sm.region.Y, sm.mx, sm.my, t)
  }
}

func (sm *StartMenu) Respond(g *gui.Gui, group gui.EventGroup) bool {
  cursor := group.Events[0].Key.Cursor()
  if cursor != nil {
    sm.mx, sm.my = cursor.Point()
  }
  return false
}

func (sm *StartMenu) Draw(region gui.Region) {
  sm.region = region
  gl.Color4ub(255, 255, 255, 255)
  sm.back.RenderNatural(sm.region.X, sm.region.Y)
  sm.menu.RenderNatural(sm.region.X+sm.layout.Menu.X, sm.region.Y+sm.layout.Menu.Y)
  // d := base.GetDictionary(sm.layout.Menu.Text.Size)
  // opts := []string{
  //   "Continue",
  //   "Versus Mode",
  //   "Story Mode",
  //   "Settings",
  // }
  // x := float64(region.X + sm.layout.Menu.X + sm.layout.Menu.Text.X)
  // y := float64(region.Y + sm.layout.Menu.Y + sm.layout.Menu.Text.Y + (len(opts)*int(d.MaxHeight()))/2 + ((len(opts)-1)*sm.layout.Menu.Text.Spacing)/2)
  // for i := range opts {
  //   d.RenderString(opts[i], x, y, 0, d.MaxHeight(), gui.Center)
  //   y -= d.MaxHeight() + float64(sm.layout.Menu.Text.Spacing)
  // }
  for _, button := range sm.buttons {
    button.RenderAt(sm.region.X, sm.region.Y)
  }
}

func (sm *StartMenu) DrawFocused(region gui.Region) {
}

func (sm *StartMenu) String() string {
  return "start menu"
}
