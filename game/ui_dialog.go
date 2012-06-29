package game

import (
  "fmt"
  "path/filepath"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/opengl/gl"
  "errors"
)

type Paragraph struct {
  X, Y, Dx, Size int
  Justification string
}

type dialogSection struct {
  // Center of the image
  X, Y int
  Paragraph Paragraph

  // The clickable region
  Region struct {
    X, Y, Dx, Dy int
  }
}

type mediumDialogLayoutSpec struct {
  Sections []dialogSection
}

type mediumDialogLayout struct {
  Background texture.Object
  Next, Prev Button

  Formats map[string]mediumDialogLayoutSpec
}

type mediumDialogData struct {
  Format string
  Sections []struct {
    Image   texture.Object
    Text    string
    shading float64
  }
}

type MediumDialogBox struct {
  layout mediumDialogLayout
  format mediumDialogLayoutSpec
  // state  mediumDialogState
  data   mediumDialogData

  region gui.Region

  buttons []*Button

  // Position of the mouse
  mx,my int

  done   bool
  result chan int
}

func MakeMediumDialogBox(source string) (*MediumDialogBox, <-chan int, error) {
  var mdb MediumDialogBox
  datadir := base.GetDataDir()
  err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "dialog", "medium.json"), "json", &mdb.layout)
  if err != nil {
    return nil, nil, err
  }
  err = base.LoadAndProcessObject(filepath.Join(datadir, source), "json", &mdb.data)
  if err != nil {
    return nil, nil, err
  }

  var ok bool
  mdb.format, ok = mdb.layout.Formats[mdb.data.Format]
  if !ok {
    return nil, nil, errors.New(fmt.Sprintf("Unknown Medium Dialog Box format '%s'.", mdb.data.Format))
  }

  if len(mdb.format.Sections) != len(mdb.data.Sections) {
    return nil, nil, errors.New(fmt.Sprintf("Format '%s' requires exactly %d sections.", mdb.data.Format, len(mdb.format.Sections)))
  }

  // return nil, nil, errors.New(fmt.Sprintf("Unknown format string: '%s'.", format))

  mdb.buttons = []*Button {
    &mdb.layout.Next,
    &mdb.layout.Prev,
  }

  mdb.result = make(chan int, 1)
  mdb.layout.Next.f = func(_data interface{}) {
    if !mdb.done {
      mdb.result <- -1
      close(mdb.result)
    }
    mdb.done = true
  }
  mdb.layout.Prev.f = func(_data interface{}) {
    if !mdb.done {
      mdb.result <- -2
      close(mdb.result)
    }
    mdb.done = true
  }

  return &mdb, mdb.result, nil
}


// type mainBarState struct {
//   Actions struct {
//     // target is the action that should be displayed as left-most,
//     // pos is the action that is currently left-most, which can be fractional.
//     scroll_target float64
//     scroll_pos    float64

//     selected Action

//     // The clicked action was clicked in the Ui but hasn't been set as the
//     // selected action yet because we can't SetCurrentAction during event
//     // handling.
//     clicked Action

//     space float64
//   }
//   Conditions struct {
//     scroll_pos float64
//   }
//   MouseOver struct {
//     active   bool
//     text     string
//     location mouseOverLocation
//   }
// }

// type mouseOverLocation int
// const (
//   mouseOverActions = iota
//   mouseOverConditions
//   mouseOverGear
// )

// func buttonFuncEndTurn(mbi interface{}) {
//   mb := mbi.(*MainBar)
//   mb.game.player_active = false
//   base.Log().Printf("player_active set to false")
// }
// func buttonFuncActionLeft(mbi interface{}) {
//   mb := mbi.(*MainBar)
//   mb.state.Actions.scroll_target -= float64(mb.layout.Actions.Count)
// }
// func buttonFuncActionRight(mbi interface{}) {
//   mb := mbi.(*MainBar)
//   mb.state.Actions.scroll_target += float64(mb.layout.Actions.Count)
// }
// func buttonFuncUnitLeft(mbi interface{}) {
//   mb := mbi.(*MainBar)
//   if !mb.game.SetCurrentAction(nil) {
//     return
//   }
//   start_index := len(mb.game.Ents) - 1
//   for i := 0; i < len(mb.game.Ents); i++ {
//     if mb.game.Ents[i] == mb.ent {
//       start_index = i
//       break
//     }
//   }
//   for i := start_index - 1; i >= 0; i-- {
//     if mb.game.Ents[i].Side() == mb.game.Side {
//       mb.game.SelectEnt(mb.game.Ents[i])
//       return
//     }
//   }
//   for i := len(mb.game.Ents) - 1; i >= start_index; i-- {
//     if mb.game.Ents[i].Side() == mb.game.Side {
//       mb.game.SelectEnt(mb.game.Ents[i])
//       return
//     }
//   }
// }
// func buttonFuncUnitRight(mbi interface{}) {
//   mb := mbi.(*MainBar)
//   if !mb.game.SetCurrentAction(nil) {
//     return
//   }
//   start_index := 0
//   for i := 0; i < len(mb.game.Ents); i++ {
//     if mb.game.Ents[i] == mb.ent {
//       start_index = i
//       break
//     }
//   }
//   for i := start_index + 1; i < len(mb.game.Ents); i++ {
//     if mb.game.Ents[i].Side() == mb.game.Side {
//       mb.game.SelectEnt(mb.game.Ents[i])
//       return
//     }
//   }
//   for i := 0; i <= start_index; i++ {
//     if mb.game.Ents[i].Side() == mb.game.Side {
//       mb.game.SelectEnt(mb.game.Ents[i])
//       return
//     }
//   }
// }

// func MakeMainBar(game *Game) (*MainBar, error) {
//   var mb MainBar
//   datadir := base.GetDataDir()
//   err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "main_bar.json"), "json", &mb.layout)
//   if err != nil {
//     return nil, err
//   }
//   mb.buttons = []*Button{
//     &mb.layout.EndTurn,
//     &mb.layout.UnitLeft,
//     &mb.layout.UnitRight,
//     &mb.layout.ActionLeft,
//     &mb.layout.ActionRight,
//   }
//   mb.layout.EndTurn.f = buttonFuncEndTurn
//   mb.layout.UnitRight.f = buttonFuncUnitRight
//   mb.layout.UnitRight.key = gin.Tab
//   mb.layout.UnitLeft.f = buttonFuncUnitLeft
//   mb.layout.UnitLeft.key = gin.ShiftTab
//   mb.layout.ActionLeft.f = buttonFuncActionLeft
//   mb.layout.ActionRight.f = buttonFuncActionRight
//   mb.game = game
//   return &mb, nil
// }
func (mdb *MediumDialogBox) Requested() gui.Dims {
  return gui.Dims{
    Dx: mdb.layout.Background.Data().Dx(),
    Dy: mdb.layout.Background.Data().Dy(),
  }
}

// func (mb *MainBar) SelectEnt(ent *Entity) {
//   if ent == mb.ent {
//     return
//   }
//   mb.ent = ent
//   mb.state = mainBarState{}

//   if mb.ent == nil {
//     return
//   }
// }

func (mdb *MediumDialogBox) Expandable() (bool, bool) {
  return false, false
}

func (mdb *MediumDialogBox) Rendered() gui.Region {
  return mdb.region
}

func (mdb *MediumDialogBox) Think(g *gui.Gui, t int64) {
  for _, button := range mdb.buttons {
    button.Think(mdb.region.X, mdb.region.Y, mdb.mx, mdb.my, t)
  }
  for i := range mdb.format.Sections {
    section := mdb.format.Sections[i]
    data := &mdb.data.Sections[i]
    if section.Region.Dx * section.Region.Dy <= 0 {
      data.shading = 1.0
    }
    in := pointInsideRect(mdb.mx, mdb.my, mdb.region.X + section.Region.X, mdb.region.Y + section.Region.Y, section.Region.Dx, section.Region.Dy)
    data.shading = doShading(data.shading, in, t)
  }
}

func (mdb *MediumDialogBox) Respond(g *gui.Gui, group gui.EventGroup) bool {
  cursor := group.Events[0].Key.Cursor()
  if cursor != nil {
    mdb.mx, mdb.my = cursor.Point()
    if !pointInsideRect(mdb.mx, mdb.my, mdb.region.X, mdb.region.Y, mdb.layout.Background.Data().Dx(), mdb.layout.Background.Data().Dy()) {
      return false
    }
  }

  for _, button := range mdb.buttons {
    if button.Respond(group, mdb) {
      return true
    }
  }

  if found, event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    for _, button := range mdb.buttons {
      if button.handleClick(mdb.mx - mdb.region.X, mdb.my - mdb.region.Y, mdb) {
        return true
      }
    }
    for i, section := range mdb.format.Sections {
      if pointInsideRect(
          mdb.mx,
          mdb.my,
          mdb.region.X + section.Region.X,
          mdb.region.Y + section.Region.Y,
          section.Region.Dx,
          section.Region.Dy) {
        if !mdb.done {
          mdb.result <- i
          close(mdb.result)
        }
        mdb.done = true
        break
      }
    }
  }

  return cursor != nil
}

func (mdb *MediumDialogBox) Draw(region gui.Region) {
  mdb.region = region
  gl.Enable(gl.TEXTURE_2D)
  gl.Color4ub(255, 255, 255, 255)
  mdb.layout.Background.Data().RenderNatural(region.X, region.Y)
  for _, button := range mdb.buttons {
    button.RenderAt(region.X, region.Y)
  }

  for i := range mdb.format.Sections {
    section := mdb.format.Sections[i]
    data := mdb.data.Sections[i]
    p := section.Paragraph
    d := base.GetDictionary(p.Size)
    var just gui.Justification
    switch p.Justification {
    case "left":
      just = gui.Left
    case "right":
      just = gui.Right
    case "center":
      just = gui.Center
    default:
      base.Error().Printf("Unknown justification '%s'", p.Justification)
      p.Justification = "left"
    }
    gl.Color4ub(255, 255, 255, 255)
    d.RenderParagraph(data.Text, float64(p.X + region.X), float64(p.Y + region.Y) - d.MaxHeight(), 0, float64(p.Dx), d.MaxHeight(), just)

    gl.Color4ub(255, 255, 255, byte(data.shading * 255))
    tex := data.Image.Data()
    tex.RenderNatural(region.X + section.X - tex.Dx() / 2 , region.Y +  section.Y - tex.Dy() / 2)
  }
}

func (mdb *MediumDialogBox) DrawFocused(region gui.Region) { }

func (mdb *MediumDialogBox) String() string {
  return "medium dialog box"
}

