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
  viewer,editor := house.MakeRoomEditorPanel(room)
  ui.AddChild(editor)
  // ui.AddChild(anch)

  sys.Think()
  ui.Draw()
  for key_map["quit"].FramePressCount() == 0 {
    sys.SwapBuffers()
    sys.Think()
    ui.Draw()
    zoom := key_map["zoom in"].FramePressAmt() - key_map["zoom out"].FramePressAmt()
    viewer.Zoom(zoom / 50)
    pan_x := key_map["pan right"].FramePressAmt() - key_map["pan left"].FramePressAmt()
    pan_y := key_map["pan up"].FramePressAmt() - key_map["pan down"].FramePressAmt()
    viewer.Move(pan_x * 7, pan_y * 7)
  }
}








