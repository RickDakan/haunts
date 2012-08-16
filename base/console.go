package base

import (
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "bufio"
  "github.com/runningwild/opengl/gl"
  "strings"
  "regexp"
)

const maxLines = 10000
const maxLineLength = 255
const maxLinesDisplayed = 30

var valid_keys map[byte]bool
var shift_keys map[byte]byte

func init() {
  valid_keys = make(map[byte]bool)
  shift_keys = make(map[byte]byte)
  keys := []byte(" abcdefghijklmnopqrstuvwxyz`1234567890-=,./;[]\\'")
  shif := []byte(" ABCDEFGHIJKLMNOPQRSTUVWXYZ~!@#$%^&*()_+<>?:{}|\"")
  for i := range keys {
    valid_keys[keys[i]] = true
    shift_keys[keys[i]] = shif[i]
  }
}

// A simple gui element that will display the last several lines of text from
// a log file (TODO: and also allow you to enter some basic commands).
type Console struct {
  gui.BasicZone
  lines       [maxLines]string
  line_buffer []string
  start, end  int
  xscroll     float64

  input   *bufio.Reader
  cmd     []byte
  dict    *gui.Dictionary
  matcher *regexp.Regexp
}

func MakeConsole() *Console {
  if log_console == nil {
    panic("Cannot make a console until the logging system has been set up.")
  }
  var c Console
  c.BasicZone.Ex = true
  c.BasicZone.Ey = true
  c.BasicZone.Request_dims = gui.Dims{1000, 1000}
  c.input = bufio.NewReader(log_console)
  c.dict = GetDictionary(12)
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

  changed := false
  for i := range group.Events {
    event := group.Events[i]
    if event.Type == gin.Press {
      r := rune(event.Key.Id())
      if r < 256 && valid_keys[byte(r)] {
        if gin.In().GetKey(gin.EitherShift).IsDown() {
          r = rune(shift_keys[byte(r)])
        }
        c.cmd = append(c.cmd, byte(r))
        changed = true
      }
    }
  }
  if found, event := group.FindEvent(gin.DeleteOrBackspace); found && event.Type == gin.Press {
    if len(c.cmd) > 0 {
      c.cmd = c.cmd[0 : len(c.cmd)-1]
      changed = true
    }
  }
  if changed {
    r, err := regexp.Compile(string(c.cmd))
    if err != nil {
      c.matcher = nil
    } else {
      c.matcher = r
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
  gl.Vertex2i(region.X, region.Y+region.Dy)
  gl.Vertex2i(region.X+region.Dx, region.Y+region.Dy)
  gl.Vertex2i(region.X+region.Dx, region.Y)
  gl.End()
  gl.Color4d(1, 1, 1, 1)
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

  c.line_buffer = c.line_buffer[0:0]
  if c.start > c.end {
    for i := c.end - 1; i >= 0; i-- {
      if len(c.line_buffer) >= maxLinesDisplayed {
        break
      }
      if c.matcher == nil || c.matcher.MatchString(c.lines[i]) {
        c.line_buffer = append(c.line_buffer, c.lines[i])
      }
    }
    for i := len(c.lines) - 1; i >= c.start; i-- {
      if len(c.line_buffer) >= maxLinesDisplayed {
        break
      }
      if c.matcher == nil || c.matcher.MatchString(c.lines[i]) {
        c.line_buffer = append(c.line_buffer, c.lines[i])
      }
    }
  } else {
    end := len(c.lines)
    if c.end < end {
      end = c.end
    }
    for i := end - 1; i >= c.start; i-- {
      if len(c.line_buffer) >= maxLinesDisplayed {
        break
      }
      if c.matcher == nil || c.matcher.MatchString(c.lines[i]) {
        c.line_buffer = append(c.line_buffer, c.lines[i])
      }
    }
  }

  y := float64(region.Y) + float64(len(c.line_buffer))*c.dict.MaxHeight()
  for i := len(c.line_buffer) - 1; i >= 0; i-- {
    do_color(c.line_buffer[i])
    c.dict.RenderString(c.line_buffer[i], c.xscroll, y, 0, c.dict.MaxHeight(), gui.Left)
    y -= c.dict.MaxHeight()
  }
  c.dict.RenderString(string(c.cmd), c.xscroll, y, 0, c.dict.MaxHeight(), gui.Left)
}
