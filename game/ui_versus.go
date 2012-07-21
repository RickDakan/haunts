package game

import (
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/util/algorithm"
  "github.com/runningwild/haunts/base"
  "path/filepath"
)

// Choose side
// Choose map
// Intruders:
//   Choose mode
//   Choose units
//   Choose gear
// Denizens:
//   Choose whatever
// Place stuff down, blizitch

func makeChooserFromOptionBasicsFile(path string) (*Chooser, <-chan map[int]bool, error) {
  var bops []OptionBasic
  err := base.LoadAndProcessObject(path, "json", &bops)
  if err != nil {
    return nil, nil, err
  }
  var opts []Option
  algorithm.Map2(bops, &opts, func(ob OptionBasic) Option { return &ob })
  return MakeChooser(opts)
}

func makeChooseGoalMenu() (*Chooser, <-chan map[int]bool, error) {
  path := filepath.Join(base.GetDataDir(), "ui", "start", "versus", "goals.json")
  return makeChooserFromOptionBasicsFile(path)
}

func makeChooseSideMenu() (*Chooser, <-chan map[int]bool, error) {
  path := filepath.Join(base.GetDataDir(), "ui", "start", "versus", "side.json")
  return makeChooserFromOptionBasicsFile(path)
}

func insertGoalMenu(ui gui.WidgetParent, replace func(gui.WidgetParent) error) error {
  chooser, done, err := makeChooseGoalMenu()
  if err != nil {
    return err
  }
  ui.AddChild(chooser)
  go func() {
    m := <-done
    ui.RemoveChild(chooser)
    if m != nil {
      base.Log().Printf("Chose: %v", m)
      err = insertGoalMenu(ui, replace)
      if err != nil {
        base.Error().Printf("Error making goal menu: %v", err)
      }
    } else {
      err := replace(ui)
      if err != nil {
        base.Error().Printf("Error replacing menu: %v", err)
      }
    }
  }()
  return nil
}

func InsertVersusMenu(ui gui.WidgetParent, replace func(gui.WidgetParent) error) error {
  chooser, done, err := makeChooseSideMenu()
  if err != nil {
    return err
  }
  ui.AddChild(chooser)
  go func() {
    m := <-done
    ui.RemoveChild(chooser)
    if m != nil {
      base.Log().Printf("Chose: %v", m)
      err = insertGoalMenu(ui, func(parent gui.WidgetParent) error {
        parent.RemoveChild(chooser)
        return InsertVersusMenu(ui, replace)
      })
      if err != nil {
        base.Error().Printf("Error making goal menu: %v", err)
      }
    } else {
      err := replace(ui)
      if err != nil {
        base.Error().Printf("Error replacing menu: %v", err)
      }
    }
  }()
  return nil
}
