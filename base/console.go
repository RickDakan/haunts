package base

import (
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "bufio"
  "github.com/runningwild/opengl/gl"
  "strings"
  "unicode"
)

const maxLines = 25
const maxLineLength = 150

// A simple gui element that will display the last several lines of text from
// a log file (TODO: and also allow you to enter some basic commands).
type Console struct {
  gui.BasicZone
  lines [maxLines]string
  start,end int
  xscroll float64

  input *bufio.Reader
  cmd []byte
}

func MakeConsole() *Console {
  if log_reader == nil {
    panic("Cannot make a console until the logging system has been set up.")
  }
  var c Console
  c.BasicZone.Ex = true
  c.BasicZone.Ey = true
  c.BasicZone.Request_dims = gui.Dims{ 1000, 1000 }
  c.input = bufio.NewReaderSize(log_reader, 1024)
  return &c
}

func (c *Console) String() string {
  return "console"
}

func (c *Console) Think(ui *gui.Gui, dt int64) {
  for line, _, err := c.input.ReadLine(); err == nil; line, _, err = c.input.ReadLine() {
    c.lines[c.end] = string(line)
    c.end = (c.end + 1) % len(c.lines)
    if c.start == c.end {
      c.start = (c.start + 1) % len(c.lines)
    }
  }
}

func (c *Console) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if found, event := group.FindEvent(GetDefaultKeyMap()["console"].Id()); found && event.Type == gin.Press {
    if group.Focus {
      ui.DropFocus()
    } else {
      ui.TakeFocus(c)
    }
    return true
  }
  if found, event := group.FindEvent(gin.Left); found && event.Type == gin.Press {
    c.xscroll += 250
  }
  if found, event := group.FindEvent(gin.Right); found && event.Type == gin.Press {
    c.xscroll -= 250
  }
  if c.xscroll > 0 {
    c.xscroll = 0
  }
  if found, event := group.FindEvent(gin.Space); found && event.Type == gin.Press {
    c.xscroll = 0
  }

  if group.Events[0].Type == gin.Press {
    r := rune(group.Events[0].Key.Id())
    if r < 256 {
      if gin.In().GetKey(gin.EitherShift).IsDown() {
        r = unicode.ToUpper(r)
      }
      c.cmd = append(c.cmd, byte(r))
    }
  }

  return group.Focus
}

func (c *Console) Draw(region gui.Region) {
}

func (c *Console) DrawFocused(region gui.Region) {
  gl.Color4d(0.2, 0, 0.3, 0.8)
  gl.Disable(gl.TEXTURE_2D)
  gl.Begin(gl.QUADS)
    gl.Vertex2i(region.X, region.Y)
    gl.Vertex2i(region.X, region.Y + region.Dy)
    gl.Vertex2i(region.X + region.Dx, region.Y + region.Dy)
    gl.Vertex2i(region.X + region.Dx, region.Y)
  gl.End()
  gl.Color4d(1, 1, 1, 1)
  d := GetDictionary(12)
  y := float64(region.Y) + float64(len(c.lines)) * d.MaxHeight()
  do_color := func(line string) {
    if strings.HasPrefix(line, "LOG") {
      gl.Color4d(1, 1, 1, 1)
    }
    if strings.HasPrefix(line, "WARN") {
      gl.Color4d(1, 1, 0, 1)
    }
    if strings.HasPrefix(line, "ERROR") {
      gl.Color4d(1, 0, 0, 1)
    }
  }
  if c.start > c.end {
    for i := c.start; i < len(c.lines); i++ {
      do_color(c.lines[i])
      d.RenderString(c.lines[i], c.xscroll, y, 0, d.MaxHeight(), gui.Left)
      y -= d.MaxHeight()
    }
    for i := 0; i < c.end; i++ {
      do_color(c.lines[i])
      d.RenderString(c.lines[i], c.xscroll, y, 0, d.MaxHeight(), gui.Left)
      y -= d.MaxHeight()
    }
  } else {
    for i := c.start; i < c.end && i < len(c.lines); i++ {
      do_color(c.lines[i])
      d.RenderString(c.lines[i], c.xscroll, y, 0, d.MaxHeight(), gui.Left)
      y -= d.MaxHeight()
    }
  }
  d.RenderString(string(c.cmd), c.xscroll, y, 0, d.MaxHeight(), gui.Left)
}
