package game

import (
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/util/algorithm"
  "github.com/runningwild/haunts/base"
  "path/filepath"
)

func InsertVersusMenu(ui gui.WidgetParent, replace func(gui.WidgetParent) error) error {
  var bops []OptionBasic
  err := base.LoadAndProcessObject(filepath.Join(base.GetDataDir(), "ui", "start", "versus", "goals.json"), "json", &bops)
  if err != nil {
    return err
  }
  var opts []Option
  algorithm.Map2(bops, &opts, func(ob OptionBasic) Option { return &ob })
  chooser, done, err := MakeChooser(opts)

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
  ui.AddChild(chooser)

  return nil
}
