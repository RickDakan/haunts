package main

import (
  "glop/gos"
  "glop/gin"
  "glop/gui"
  "glop/system"
  "runtime"
  "path/filepath"
  "os"
)

var (
  sys system.System
  datadir string
)

func init() {
  runtime.LockOSThread()
  sys = system.Make(gos.GetSystemInterface())

  // TODO: This should not be OS-specific
  datadir = filepath.Join(os.Args[0], "..", "..")
}

func main() {
  sys.Startup()
  factor := 0.92
  wdx := int(1024 * factor)
  wdy := int(768 * factor)
  sys.CreateWindow(10, 10, wdx, wdy)
  sys.EnableVSync(true)
  ui,err := gui.Make(gin.In(), gui.Dims{ wdx, wdy }, filepath.Join(datadir, "fonts", "skia.ttf"))
  if err != nil {
    panic(err.Error())
  }
  for gin.In().GetKey('q').FramePressCount() == 0 {
    sys.Think()
    ui.Draw()
  }
}
