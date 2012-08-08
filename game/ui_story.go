package game

import (
  "path/filepath"
  "sort"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/opengl/gl"
)

type storyLayout struct {
  Title struct {
    X, Y    int
    Texture texture.Object
  }
  Background texture.Object
  Back       Button
  New        Button
  Continue   Button
}

type StoryMenu struct {
  layout  storyLayout
  region  gui.Region
  buttons []*Button
  mx, my  int
}

func InsertStoryMenu(ui gui.WidgetParent) error {
  var sm StoryMenu
  datadir := base.GetDataDir()
  err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "start", "story", "layout.json"), "json", &sm.layout)
  if err != nil {
    return err
  }
  sm.buttons = []*Button{
    &sm.layout.Back,
    &sm.layout.New,
    // &sm.layout.Continue,
  }
  sm.layout.Back.f = func(interface{}) {
    ui.RemoveChild(&sm)
    InsertStartMenu(ui)
  }
  sm.layout.New.f = func(interface{}) {
    ui.RemoveChild(&sm)
    ui.AddChild(MakeGamePanel("foo.lua", &Player{}, nil))
  }

  players := GetAllPlayers()
  var player_names []string
  for player_name := range players {
    player_names = append(player_names, player_name)
  }
  sort.Strings(player_names)
  y := sm.layout.Continue.Y
  for i := range player_names {
    player_name := player_names[i]
    var button Button
    button.X = sm.layout.Continue.X
    button.Y = y
    y -= sm.layout.New.Y - sm.layout.Continue.Y
    button.Text.String = player_name
    button.Text.Size = sm.layout.Continue.Text.Size
    button.Text.Justification = sm.layout.Continue.Text.Justification
    button.f = func(interface{}) {
      ui.RemoveChild(&sm)
      p, err := LoadPlayer(players[player_name])
      if err != nil {
        base.Warn().Printf("Failed to load player '%s': %v", player_name, err)
      }
      ui.AddChild(MakeGamePanel("", p, nil))
      base.Log().Printf("Pressed %s", player_name)
    }
    sm.buttons = append(sm.buttons, &button)
  }

  // sm.layout.Continue.f = func(interface{}) {}
  ui.AddChild(&sm)
  return nil
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
  for _, button := range sm.buttons {
    button.RenderAt(sm.region.X, sm.region.Y)
  }
}

func (sm *StoryMenu) DrawFocused(region gui.Region) {
}

func (sm *StoryMenu) String() string {
  return "story menu"
}
