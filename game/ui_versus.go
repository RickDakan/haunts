package game

import (
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/opengl/gl"
  "path/filepath"
)

type versusLayout struct {
  Title struct {
    X, Y    int
    Texture texture.Object
  }
  Background texture.Object
  Back       Button
  New        Button
  Continue   Button
}

type VersusMenu struct {
  layout  storyLayout
  region  gui.Region
  buttons []*Button
  mx, my  int
}

func InsertVersusMenu(ui gui.WidgetParent, replace func(gui.WidgetParent) error) error {
  var vm VersusMenu
  datadir := base.GetDataDir()
  err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "start", "versus", "layout.json"), "json", &vm.layout)
  if err != nil {
    return err
  }
  vm.buttons = []*Button{
    &vm.layout.Back,
    &vm.layout.New,
    &vm.layout.Continue,
  }
  vm.layout.Back.f = func(interface{}) {
    ui.RemoveChild(&vm)
    InsertStartMenu(ui)
  }
  vm.layout.New.f = func(interface{}) {
    ui.RemoveChild(&vm)
    ui.AddChild(MakeGamePanel())
  }
  vm.layout.Continue.f = func(interface{}) {}
  chooser, done, err := MakeChooser()
  if err != nil {
    return err
  }
  go func() {
    m := <-done
    ui.RemoveChild(chooser)
    if m == nil {
      err := replace(ui)
      if err != nil {
        base.Error().Printf("Error replacing menu: %v", err)
      }
    } else {
      err := replace(ui)
      if err != nil {
        base.Error().Printf("Error replacing menu: %v", err)
      }
    }
  }()
  // ui.AddChild(&vm)
  ui.AddChild(chooser)
  return nil
}

func (vm *VersusMenu) Requested() gui.Dims {
  return gui.Dims{1024, 768}
}

func (vm *VersusMenu) Expandable() (bool, bool) {
  return false, false
}

func (vm *VersusMenu) Rendered() gui.Region {
  return vm.region
}

func (vm *VersusMenu) Think(g *gui.Gui, t int64) {
  if vm.mx == 0 && vm.my == 0 {
    vm.mx, vm.my = gin.In().GetCursor("Mouse").Point()
  }
  for _, button := range vm.buttons {
    button.Think(vm.region.X, vm.region.Y, vm.mx, vm.my, t)
  }
}

func (vm *VersusMenu) Respond(g *gui.Gui, group gui.EventGroup) bool {
  cursor := group.Events[0].Key.Cursor()
  if cursor != nil {
    vm.mx, vm.my = cursor.Point()
  }
  if found, event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    for _, button := range vm.buttons {
      if button.handleClick(vm.mx, vm.my, nil) {
        return true
      }
    }
  }
  return false
}

func (vm *VersusMenu) Draw(region gui.Region) {
  vm.region = region
  gl.Color4ub(255, 255, 255, 255)
  vm.layout.Background.Data().RenderNatural(region.X, region.Y)
  title := vm.layout.Title
  title.Texture.Data().RenderNatural(region.X+title.X, region.Y+title.Y)
  for _, button := range vm.buttons {
    button.RenderAt(vm.region.X, vm.region.Y)
  }
}

func (vm *VersusMenu) DrawFocused(region gui.Region) {
}

func (vm *VersusMenu) String() string {
  return "start menu"
}
