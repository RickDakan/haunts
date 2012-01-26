package main

import (
  "glop/gos"
  "glop/gin"
  "glop/gui"
  "glop/system"
  "glop/render"
  "runtime"
  "path/filepath"
  "os"
  "fmt"
  "runtime/debug"
  "haunts/house"
  "haunts/base"
  "haunts/game"
)

var (
  sys system.System
  datadir string
  key_map base.KeyMap
  editors map[string]house.Editor
  editor house.Editor
  editor_name string
  ui *gui.Gui
  anchor *gui.AnchorBox
  chooser *gui.FileChooser
  wdx,wdy int
  game_panel *game.GamePanel
  zooming,dragging,hiding bool
)

func loadAllRegistries() {
  house.LoadAllFurnitureInDir(filepath.Join(datadir, "furniture"))
  house.LoadAllWallTexturesInDir(filepath.Join(datadir, "textures"))
  house.LoadAllRoomsInDir(filepath.Join(datadir, "rooms"))
  house.LoadAllDoorsInDir(filepath.Join(datadir, "doors"))
  house.LoadAllHousesInDir(filepath.Join(datadir, "houses"))
}

func init() {
  runtime.LockOSThread()
  sys = system.Make(gos.GetSystemInterface())

  var key_binds base.KeyBinds
  base.LoadJson("/Users/runningwild/code/haunts/key_binds.json", &key_binds)
  key_map = key_binds.MakeKeyMap()
  base.SetDefaultKeyMap(key_map)

  // TODO: This should not be OS-specific
  datadir = filepath.Join(os.Args[0], "..", "..")
  base.SetDatadir(datadir)
  err := house.SetDatadir(datadir)
  if err != nil {
    panic(err.Error())
  }

  factor := 1.0
  wdx = int(1200 * factor)
  wdy = int(675 * factor)
}

type draggerZoomer interface {
  Drag(float64, float64)
  Zoom(float64)
}

func draggingAndZooming(dz draggerZoomer) {
  if ui.FocusWidget() != nil {
    dragging = false
    zooming = false
    sys.HideCursor(false)
    return
  }

  if key_map["zoom"].IsDown() != zooming {
    zooming = !zooming
  }
  if zooming {
    zoom := gin.In().GetKey(gin.MouseWheelVertical).FramePressAmt()
    dz.Zoom(zoom / 100)
  }

  if key_map["drag"].IsDown() != dragging {
    dragging = !dragging
  }
  if dragging {
    mx := gin.In().GetKey(gin.MouseXAxis).FramePressAmt()
    my := gin.In().GetKey(gin.MouseYAxis).FramePressAmt()
    if mx != 0 || my != 0 {
      dz.Drag(-mx, my)
    }
  }

  if (dragging || zooming) != hiding {
    hiding = (dragging || zooming)
    sys.HideCursor(hiding)
  }
}

func gameMode() {
  draggingAndZooming(game_panel.GetViewer())
}

func editMode() {
  draggingAndZooming(editor.GetViewer())
  if ui.FocusWidget() == nil {
    for name := range editors {
      if key_map[fmt.Sprintf("%s editor", name)].FramePressCount() > 0 && ui.FocusWidget() == nil {
        ui.RemoveChild(editor)
        editor_name = name
        editor = editors[editor_name]
        loadAllRegistries()
        editor.Reload()
        ui.AddChild(editor)
      }
    }

    if key_map["save"].FramePressCount() > 0 && chooser == nil {
      path,err := editor.Save()
      if path != "" && err == nil {
        base.SetStoreVal(fmt.Sprintf("last %s path", editor_name), base.TryRelative(datadir, path))
      }
    }

    if key_map["load"].FramePressCount() > 0 && chooser == nil {
      callback := func(path string, err error) {
        ui.DropFocus()
        ui.RemoveChild(anchor)
        chooser = nil
        anchor = nil
        err = editor.Load(path)
        if err != nil {
        } else {
          base.SetStoreVal(fmt.Sprintf("last %s path", editor_name), base.TryRelative(datadir, path))
        }
      }
      chooser = gui.MakeFileChooser(filepath.Join(datadir, fmt.Sprintf("%ss", editor_name)), callback, gui.MakeFileFilter(fmt.Sprintf(".%s", editor_name)))
      anchor = gui.MakeAnchorBox(gui.Dims{ wdx, wdy })
      anchor.AddChild(chooser, gui.Anchor{ 0.5, 0.5, 0.5, 0.5 })
      ui.AddChild(anchor)
      ui.TakeFocus(chooser)
    }


    // Don't select tabs in an editor if we're doing some other sort of command
    ok_to_select := true
    for _,v := range key_map {
      if v.FramePressCount() > 0 {
        ok_to_select = false
        break
      }
    }
    if ok_to_select {
      for i := 1; i <= 9; i++ {
        if gin.In().GetKey(gin.KeyId('0' + i)).FramePressCount() > 0 {
          editor.SelectTab(i - 1)
        }
      }
    }
  }
}

func main() {
  defer func() {
    if r := recover(); r != nil {
      data := debug.Stack()
      fmt.Printf("Panic: %v\n", r)
      fmt.Printf("%s\n", string(data))
      out,err := os.Open("crash.txt")
      if err == nil {
        out.Write([]byte(fmt.Sprintf("Panic: %v\n", r)))
        out.Write(data)
        out.Close()
      }
    }
  } ()
  sys.Startup()
  render.Init()
  render.Queue(func() {
    sys.CreateWindow(10, 10, wdx, wdy)
    sys.EnableVSync(true)
  })
  var err error
  ui,err = gui.Make(gin.In(), gui.Dims{ wdx, wdy }, filepath.Join(datadir, "fonts", "skia.ttf"))
  if err != nil {
    panic(err.Error())
  }
  loadAllRegistries()

  // TODO: Might want to be able to reload stuff, but this is sensitive because it
  // is loading textures.  We should probably redo the sprite system so that this
  // is easier to safely handle.
  game.LoadAllEntitiesInDir(filepath.Join(datadir, "entities"))

  // Set up editors
  editors = map[string]house.Editor {
    "room" : house.MakeRoomEditorPanel(),
    "house" : house.MakeHouseEditorPanel(),
  }
  for name,editor := range editors {
    path := base.GetStoreVal(fmt.Sprintf("last %s path", name))
    path = filepath.Join(datadir, path)
    if path != "" {
      editor.Load(path)
    }
  }
  editor_name = "house"
  editor = editors[editor_name]
  game_panel = game.MakeGamePanel()
  game_panel.LoadHouse("name")

  ui.AddChild(editor)
  sys.Think()
  render.Queue(func() {
    ui.Draw()
  })
  render.Purge()
  runtime.GOMAXPROCS(8)

  edit_mode := true

  for key_map["quit"].FramePressCount() == 0 {
    sys.Think()
    render.Queue(func() {
      sys.SwapBuffers()
      ui.Draw()
    })
    render.Purge()

    if key_map["game mode"].FramePressCount() % 2 == 1 {
      if edit_mode {
        ui.RemoveChild(editor)
        ui.AddChild(game_panel)
      } else {
        ui.RemoveChild(game_panel)
        ui.AddChild(editor)
      }
      edit_mode = !edit_mode
    }

    if edit_mode {
      editMode()
    } else {
      gameMode()
    }
  }
}
