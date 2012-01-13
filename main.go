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
)

var (
  sys system.System
  datadir string
  key_map base.KeyMap
  quit gin.Key
)

func loadAllRegistries() {
  house.LoadAllFurnitureInDir(filepath.Join(datadir, "furniture"))
  house.LoadAllWallTexturesInDir(filepath.Join(datadir, "textures"))
  house.LoadAllRoomsInDir(filepath.Join(datadir, "rooms"))
  house.LoadAllDoorsInDir(filepath.Join(datadir, "doors"))
}

func init() {
  runtime.LockOSThread()
  sys = system.Make(gos.GetSystemInterface())

  var key_binds base.KeyBinds
  base.LoadJson("/Users/runningwild/code/haunts/key_binds.json", &key_binds)
  key_map = key_binds.MakeKeyMap()
  base.SetDefaultKeyMap(key_map)

  quit = gin.In().BindDerivedKey("quit", gin.In().MakeBinding('q', []gin.KeyId{ gin.EitherShift }, []bool{ true }))
  // TODO: This should not be OS-specific
  datadir = filepath.Join(os.Args[0], "..", "..")
  base.SetDatadir(datadir)
  err := house.SetDatadir(datadir)
  if err != nil {
    panic(err.Error())
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
  factor := 1.0
  wdx := int(1200 * factor)
  wdy := int(675 * factor)
  render.Init()
  render.Queue(func() {
    sys.CreateWindow(10, 10, wdx, wdy)
    sys.EnableVSync(true)
  })
  ui,err := gui.Make(gin.In(), gui.Dims{ wdx, wdy }, filepath.Join(datadir, "fonts", "skia.ttf"))
  if err != nil {
    panic(err.Error())
  }
  loadAllRegistries()

  var editor house.Editor
  editor = house.MakeRoomEditorPanel()
  path := base.GetStoreVal("last room path")
  if path != "" {
    editor.Load(path)
  }
  ui.AddChild(editor)

  sys.Think()
  render.Queue(func() {
    ui.Draw()
  })
  render.Purge()
  runtime.GOMAXPROCS(8)
  var anchor *gui.AnchorBox
  var chooser *gui.FileChooser
  zooming := false
  dragging := false
  hiding := false

  // cancel is used to remove a modal gui widget
  var cancel func()

  for key_map["quit"].FramePressCount() == 0 {
    sys.SwapBuffers()
    sys.Think()
    render.Queue(func() {
      ui.Draw()
    })
    render.Purge()
    if ui.FocusWidget() != nil {
      dragging = false
      zooming = false
      sys.HideCursor(false)
      if gin.In().GetKey(gin.Escape).FramePressCount() > 0 && cancel != nil {
        cancel()
        cancel = nil
      }
    }
    if ui.FocusWidget() == nil {
      if key_map["zoom"].IsDown() != zooming {
        zooming = !zooming
      }
      if zooming {
        zoom := gin.In().GetKey(gin.MouseWheelVertical).FramePressAmt()
        editor.GetViewer().Zoom(zoom / 100)
      }

      if key_map["drag"].IsDown() != dragging {
        dragging = !dragging
      }
      if dragging {
        mx := gin.In().GetKey(gin.MouseXAxis).FramePressAmt()
        my := gin.In().GetKey(gin.MouseYAxis).FramePressAmt()
        if mx != 0 || my != 0 {
          editor.GetViewer().Drag(-mx, my)
        }
      }

      if (dragging || zooming) != hiding {
        hiding = (dragging || zooming)
        sys.HideCursor(hiding)
      }

      if key_map["editor"].FramePressCount() > 0 && ui.FocusWidget() == nil {
        vtable := gui.MakeVerticalTable()
        funcs := []struct{ text string; f func() house.Editor } {
          {
            "Room Editor",
            func() house.Editor {
              path := base.GetStoreVal("last room path")
              if path != "" {
                return house.MakeRoomEditorPanel()
              }
              return house.MakeRoomEditorPanel()
            },
          },
          {
            "House Editor",
            func() house.Editor {
              path := base.GetStoreVal("last house path")
              if path != "" {
                return house.MakeHouseEditorPanel(house.LoadHouseDef(path), datadir)
              }
              return house.MakeHouseEditorPanel(house.MakeHouseDef(), datadir)
            },
          },
        }
        for _,temp_f := range funcs {
          f := temp_f
          vtable.AddChild(gui.MakeButton("standard", f.text, 300, 1, 1, 1, 1, func(int64) {
            ui.RemoveChild(vtable)
            ui.DropFocus()
            ui.RemoveChild(editor)
            loadAllRegistries()
            editor = f.f()
            ui.AddChild(editor)
          }))
        }
        ui.AddChild(vtable)
        ui.TakeFocus(vtable)
        cancel = func() {
          ui.RemoveChild(vtable)
          ui.DropFocus()
        }
      }

      if key_map["save"].FramePressCount() > 0 && chooser == nil {
        path,err := editor.Save()
        if path != "" && err == nil {
          base.SetStoreVal("last room path", path)
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
            base.SetStoreVal("last room path", path)
          }
        }
        chooser = gui.MakeFileChooser(datadir, callback, gui.MakeFileFilter(".room"))
        anchor = gui.MakeAnchorBox(gui.Dims{ wdx, wdy })
        anchor.AddChild(chooser, gui.Anchor{ 0.5, 0.5, 0.5, 0.5 })
        ui.AddChild(anchor)
        ui.TakeFocus(chooser)
      }

      for i := 1; i <= 9; i++ {
        if gin.In().GetKey(gin.KeyId('0' + i)).FramePressCount() > 0 {
          editor.SelectTab(i - 1)
        }
      }

    }
  }
}
