package game

import (
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/opengl/gl"
  "path/filepath"
)

type storyLayout struct {
  Title struct {
    X, Y    int
    Texture texture.Object
  }
  Background texture.Object
  Back       Button
}

type StoryMenu struct {
  parent  *StartMenu
  layout  storyLayout
  region  gui.Region
  buttons []*Button
  mx, my  int
}

func MakeStoryMenu(parent *StartMenu) (*StoryMenu, error) {
  var sm StoryMenu
  datadir := base.GetDataDir()
  err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "start", "story", "layout.json"), "json", &sm.layout)
  if err != nil {
    return nil, err
  }
  sm.buttons = []*Button{
    &sm.layout.Back,
  }
  sm.layout.Back.f = func(interface{}) {
    parent.sub_menu = nil
  }
  return &sm, nil
}

func (sm *StoryMenu) Requested() gui.Dims {
  return gui.Dims{1024, 768}
}

func (sm *StoryMenu) Expandable() (bool, bool) {
  return false, false
}

func (sm *StoryMenu) Rendered() gui.Region {
  return sm.region
}

func (sm *StoryMenu) Think(g *gui.Gui, t int64) {
  if sm.mx == 0 && sm.my == 0 {
    sm.mx, sm.my = gin.In().GetCursor("Mouse").Point()
  }
  for _, button := range sm.buttons {
    button.Think(sm.region.X, sm.region.Y, sm.mx, sm.my, t)
  }
}

func (sm *StoryMenu) Respond(g *gui.Gui, group gui.EventGroup) bool {
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

func (sm *StoryMenu) Draw(region gui.Region) {
  sm.region = region
  gl.Color4ub(255, 255, 255, 255)
  sm.layout.Background.Data().RenderNatural(region.X, region.Y)
  title := sm.layout.Title
  title.Texture.Data().RenderNatural(region.X+title.X, region.Y+title.Y)
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

func (sm *StoryMenu) DrawFocused(region gui.Region) {
}

func (sm *StoryMenu) String() string {
  return "start menu"
}
