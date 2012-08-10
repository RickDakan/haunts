package game

import (
  // "fmt"
  "path/filepath"
  // "sort"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/opengl/gl"
)

type systemLayout struct {
  Main Button
  Sub  struct {
    Background texture.Object
  }
}

type SystemMenu struct {
  layout  systemLayout
  region  gui.Region
  buttons []*Button
  mx, my  int
  last_t  int64
  focus   bool
}

func MakeSystemMenu() (gui.Widget, error) {
  var sm SystemMenu
  datadir := base.GetDataDir()
  err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "system", "layout.json"), "json", &sm.layout)
  if err != nil {
    return nil, err
  }

  sm.layout.Main.f = func(interface{}) {}

  return &sm, nil
}

func (sm *SystemMenu) Requested() gui.Dims {
  return gui.Dims{1024, 768}
}

func (sm *SystemMenu) Expandable() (bool, bool) {
  return false, false
}

func (sm *SystemMenu) Rendered() gui.Region {
  return sm.region
}

func (sm *SystemMenu) Think(g *gui.Gui, t int64) {
  if sm.last_t == 0 {
    sm.last_t = t
    return
  }
  dt := t - sm.last_t
  sm.last_t = t
  if sm.mx == 0 && sm.my == 0 {
    sm.mx, sm.my = gin.In().GetCursor("Mouse").Point()
  }
  if sm.focus {
    for _, button := range sm.buttons {
      button.Think(sm.region.X, sm.region.Y, sm.mx, sm.my, dt)
    }
  } else {
    sm.layout.Main.Think(sm.region.X, sm.region.Y, sm.mx, sm.my, dt)
  }
  sm.focus = (g.FocusWidget() == sm)
}

func (sm *SystemMenu) Respond(g *gui.Gui, group gui.EventGroup) bool {
  cursor := group.Events[0].Key.Cursor()
  if cursor != nil {
    sm.mx, sm.my = cursor.Point()
  }
  if found, event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    if sm.layout.Main.handleClick(sm.mx, sm.my, nil) {
      if sm.focus {
        g.DropFocus()
      } else {
        g.TakeFocus(sm)
      }
      sm.focus = true
      base.Log().Printf("focus: %v %v", sm, g.FocusWidget())
      return true
    }
  }
  return false
}

func (sm *SystemMenu) Draw(region gui.Region) {
  sm.region = region
  gl.Color4ub(255, 255, 255, 255)
  x := region.X + region.Dx - sm.layout.Main.Texture.Data().Dx()
  y := region.Y + region.Dy - sm.layout.Main.Texture.Data().Dy()
  sm.layout.Main.RenderAt(x, y)
}

func (sm *SystemMenu) DrawFocused(region gui.Region) {
  sm.region = region
  gl.Color4ub(255, 255, 255, 255)
  x := region.X + region.Dx/2 - sm.layout.Sub.Background.Data().Dx()/2
  y := region.Y + region.Dy/2 - sm.layout.Sub.Background.Data().Dy()/2
  sm.layout.Sub.Background.Data().RenderNatural(x, y)
  for _, button := range sm.buttons {
    button.RenderAt(x, y)
  }
}

func (sm *SystemMenu) String() string {
  return "system menu"
}
