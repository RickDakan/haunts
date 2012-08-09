package game

import (
  "fmt"
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
  Text       struct {
    String        string
    Size          int
    Justification string
  }
  Up      Button
  Down    Button
  Options ScrollingRegion
}

type StoryMenu struct {
  layout             storyLayout
  region             gui.Region
  buttons            []*Button
  option_buttons     []*Button
  non_option_buttons []*Button
  mx, my             int
  last_t             int64
}

func InsertStoryMenu(ui gui.WidgetParent) error {
  var sm StoryMenu
  datadir := base.GetDataDir()
  err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "start", "story", "layout.json"), "json", &sm.layout)
  if err != nil {
    return err
  }
  sm.non_option_buttons = []*Button{
    &sm.layout.Back,
    &sm.layout.Up,
    &sm.layout.Down,
  }
  sm.layout.Back.f = func(interface{}) {
    ui.RemoveChild(&sm)
    InsertStartMenu(ui)
  }
  sm.layout.Up.f = func(interface{}) {
    sm.layout.Options.Up()
  }
  sm.layout.Down.f = func(interface{}) {
    sm.layout.Options.Down()
  }

  players := GetAllPlayers()
  var player_names []string
  for player_name := range players {
    player_names = append(player_names, player_name)
  }
  for i := 0; i < 10; i++ {
    player_names = append(player_names, fmt.Sprintf("Player %d", i))
  }
  sort.Strings(player_names)
  line_height := int(base.GetDictionary(sm.layout.Text.Size).MaxHeight())
  y := -line_height
  for i := range player_names {
    player_name := player_names[i]
    var button Button
    button.X = sm.layout.Options.X
    button.Y = y
    y -= line_height
    button.Text = sm.layout.Text
    button.Text.String = player_name
    button.f = func(interface{}) {
      ui.RemoveChild(&sm)
      p, err := LoadPlayer(players[player_name])
      if err != nil {
        base.Warn().Printf("Failed to load player '%s': %v", player_name, err)
      }
      ui.AddChild(MakeGamePanel("", p, nil))
      base.Log().Printf("Pressed %s", player_name)
    }
    sm.option_buttons = append(sm.option_buttons, &button)
  }
  sm.layout.Options.Height = line_height * len(sm.option_buttons)
  base.Log().Printf("Num elements: %d", len(sm.option_buttons))
  for _, b := range sm.non_option_buttons {
    sm.buttons = append(sm.buttons, b)
  }
  for _, b := range sm.option_buttons {
    sm.buttons = append(sm.buttons, b)
  }
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
  if sm.last_t == 0 {
    sm.last_t = t
    return
  }
  dt := t - sm.last_t
  sm.last_t = t
  sm.layout.Options.Think(dt)
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
  sm.layout.Options.Region().PushClipPlanes()
  for _, button := range sm.option_buttons {
    button.RenderAt(sm.region.X, sm.region.Y+sm.layout.Options.Top())
  }
  sm.layout.Options.Region().PopClipPlanes()
  for _, button := range sm.non_option_buttons {
    button.RenderAt(sm.region.X, sm.region.Y)
  }
}

func (sm *StoryMenu) DrawFocused(region gui.Region) {
}

func (sm *StoryMenu) String() string {
  return "story menu"
}
