package main

import (
  "glop/gos"
  "glop/gin"
  "glop/gui"
  "glop/system"
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


func init() {
  runtime.LockOSThread()
  sys = system.Make(gos.GetSystemInterface())

  var key_binds base.KeyBinds
  base.LoadJson("/Users/runningwild/code/haunts/key_binds.json", &key_binds)
  key_map = key_binds.MakeKeyMap()

  quit = gin.In().BindDerivedKey("quit", gin.In().MakeBinding('q', []gin.KeyId{ gin.EitherShift }, []bool{ true }))
  // TODO: This should not be OS-specific
  datadir = filepath.Join(os.Args[0], "..", "..")
  err := house.SetDatadir(datadir)
  if err != nil {
    panic(err.Error())
  }
}

func main() {
  defer func() {
    if r := recover(); r != nil {
      data := debug.Stack()
      fmt.Printf("%s\n", string(data))
      out,err := os.Open("crash.txt")
      if err != nil {
        out.Write(data)
        out.Close()
      }
    }
  } ()
  sys.Startup()
  factor := 1.0
  wdx := int(1200 * factor)
  wdy := int(675 * factor)
  sys.CreateWindow(10, 10, wdx, wdy)
  sys.EnableVSync(true)
  ui,err := gui.Make(gin.In(), gui.Dims{ wdx, wdy }, filepath.Join(datadir, "fonts", "skia.ttf"))
  if err != nil {
    panic(err.Error())
  }
  // anch := gui.MakeAnchorBox(gui.Dims{ wdx, wdy })
  room := house.MakeRoom()
  // anch.AddChild(house.MakeRoomEditorPanel(room), gui.Anchor{ 0.5, 0.5, 0.5, 0.5})
  editor := house.MakeRoomEditorPanel(room, datadir)
  viewer := editor.RoomViewer
  ui.AddChild(editor)
  // ui.AddChild(anch)

  sys.Think()
  ui.Draw()
  runtime.GOMAXPROCS(8)
  var anchor *gui.AnchorBox
  var chooser *gui.FileChooser
  for key_map["quit"].FramePressCount() == 0 {
    sys.SwapBuffers()
    sys.Think()
    ui.Draw()
    zoom := key_map["zoom in"].FramePressSum() - key_map["zoom out"].FramePressSum()
    viewer.Zoom(zoom / 500)
    pan_x := key_map["pan right"].FramePressSum() - key_map["pan left"].FramePressSum()
    pan_y := key_map["pan up"].FramePressSum() - key_map["pan down"].FramePressSum()
    if key_map["load"].FramePressCount() > 0 && chooser == nil {
      callback := func(path string, err error) {
        if err != nil && filepath.Ext(path) == ".room" {
          // Load room
        }
        ui.DropFocus()
        ui.RemoveChild(anchor)
        chooser = nil
        anchor = nil

        new_room := house.LoadRoom(path)
        if new_room != nil {
          ui.RemoveChild(editor)
          room = new_room
          editor = house.MakeRoomEditorPanel(room, datadir)
          viewer = editor.RoomViewer
          ui.AddChild(editor)
        }
      }
      chooser = gui.MakeFileChooser(datadir, callback, gui.MakeFileFilter(".room"))
      anchor = gui.MakeAnchorBox(gui.Dims{ wdx, wdy })
      anchor.AddChild(chooser, gui.Anchor{ 0.5, 0.5, 0.5, 0.5 })
      ui.AddChild(anchor)
      ui.TakeFocus(chooser)
    }

    viewer.Move(pan_x, pan_y)
  }
}








