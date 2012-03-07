package main

import (
  "fmt"
  "os"
  "path/filepath"
  "runtime"
  "runtime/debug"
  "runtime/pprof"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gos"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/render"
  "github.com/runningwild/glop/system"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  "github.com/runningwild/haunts/house"

  // Need to pull in all of the actions we define here and not in
  // haunts/game because haunts/game/actions depends on it
  _ "github.com/runningwild/haunts/game/actions"
  _ "github.com/runningwild/haunts/game/ai"

  "github.com/runningwild/haunts/game/status"
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
  game.RegisterActions()
  status.RegisterAllConditions()
}

func init() {
  runtime.LockOSThread()
  sys = system.Make(gos.GetSystemInterface())

  // TODO: This should not be OS-specific
  datadir = filepath.Join(os.Args[0], "..", "..")
  base.SetDatadir(datadir)
  base.Log().Printf("Setting datadir: %s", datadir)
  err := house.SetDatadir(datadir)
  if err != nil {
    panic(err.Error())
  }

  var key_binds base.KeyBinds
  base.LoadJson(filepath.Join(datadir, "key_binds.json"), &key_binds)
  key_map = key_binds.MakeKeyMap()
  base.SetDefaultKeyMap(key_map)

  wdx = 1024
  wdy = 768
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
  if key_map["load"].FramePressCount() > 0 && chooser == nil {
    callback := func(path string, err error) {
      ui.DropFocus()
      ui.RemoveChild(anchor)
      chooser = nil
      anchor = nil
      game_panel.LoadHouse(path)
      base.SetStoreVal("last game path", base.TryRelative(datadir, path))
    }
    chooser = gui.MakeFileChooser(filepath.Join(datadir, "houses"), callback, gui.MakeFileFilter(fmt.Sprintf(".house")))
    anchor = gui.MakeAnchorBox(gui.Dims{ wdx, wdy })
    anchor.AddChild(chooser, gui.Anchor{ 0.5, 0.5, 0.5, 0.5 })
    ui.AddChild(anchor)
    ui.TakeFocus(chooser)
  }
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
      if err != nil {
        base.Warn().Printf("Failed to save: %v", err.Error())
      }
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
          base.Warn().Printf("Failed to load: %v", err.Error())
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
      base.Error().Printf("PANIC: %s\n", string(data))
    }
  } ()
  base.Log().Printf("Version %s", Version())
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
  game_panel.LoadHouse(filepath.Join(datadir, base.GetStoreVal("last game path")))

  ui.AddChild(editor)
  sys.Think()
  // Wait until now to create the dictionary because the render thread needs
  // to be running in advance.
  render.Queue(func() {
    sys.Think()
    ui.Draw()
  })
  render.Purge()
  runtime.GOMAXPROCS(8)

  edit_mode := true

  var profile_output *os.File

  for key_map["quit"].FramePressCount() == 0 {
    render.Queue(func() {
      sys.Think()
      sys.SwapBuffers()
      ui.Draw()
    })
    render.Purge()

    if key_map["profile"].FramePressCount() > 0 {
      if profile_output == nil {
        profile_output, err = os.Create(filepath.Join(datadir, "cpu.prof"))
        if err == nil {
          err = pprof.StartCPUProfile(profile_output)
          if err != nil {
            base.Log().Printf("Unable to start CPU profile: %v\n", err)
            profile_output.Close()
            profile_output = nil
          }
          base.Log().Printf("profout: %v\n", profile_output)
        } else {
          base.Log().Printf("Unable to start CPU profile: %v\n", err)
        }
      } else {
        pprof.StopCPUProfile()
        profile_output.Close()
        profile_output = nil
      }
    }

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
