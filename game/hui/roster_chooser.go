package hui  // haunts ui

import (
  "github.com/runningwild/opengl/gl"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/gin"
  "path/filepath"
)

type Option interface {
  // index is the index of this option into the layout's array of options,
  // and is also the index into the map selected.  hovered indicates whether
  // or not the mouse is over this particular option.  selected is a map from
  // index to hether or not that option is selected right now.
  Draw(hovered, selected, selectable bool, region gui.Region)
  Think(dt int64)
}

type RosterChooserLayout struct {
  Num_options int
  Option struct {
    Dx, Dy int
  }
  Up, Down texture.Object

  // speed at which the scrolling happens
  // 0.0 doesn't move at all
  // 1.0 is instantaneous
  Speed float64

  // Might want background too?  Mabye textures for the other buttons?
}

type RosterChooser struct {
  gui.BasicZone
  layout  RosterChooserLayout
  options []Option

  // last position of the mouse cursor
  mouse gui.Point

  // So we can give a dt to the options if they want to animate
  last_think int64

  // What option is at the top of the list
  focus int

  // As we move the focus around we gradually move our view to smoothly
  // adjust
  focus_pos float64

  // What options we have currently selected
  selected map[int]bool

  selection_filter SelectionFilter

  on_complete func(map[int]bool)

  // Render regions - makes it easy to remember where we rendered things so we
  // know where to check for clicks.
  render struct {
    up, down    gui.Region
    options     []gui.Region
    all_options gui.Region
    done, sure  gui.Region
  }
}

// A SelectionFilter is a function that, given the index of the option the
// user is trying to select and the set of options already selected, returns
// true iff the user is allowed to select that option.  This is used to
// determine if an option qualifies for 'hovered' as well as if it can be
// selected.
// If an index of -1 is passed in this indicates that the user is trying to
// complete their selection, in this case the function should return true
// if the selection is valid.
type SelectionFilter func(int, map[int]bool) bool

func SelectAtMostN(n int) SelectionFilter {
  return func(k int, m map[int]bool) bool {
    if k == -1 {
      return len(m) <= n
    }
    return m[k] || len(m) < n
  }
}

func MakeRosterChooser(options []Option, filter SelectionFilter, on_complete func(map[int]bool)) *RosterChooser {
  var rc RosterChooser
  rc.options = options
  err := base.LoadAndProcessObject(filepath.Join(base.GetDataDir(), "ui", "widgets", "roster_chooser.json"), "json", &rc.layout)
  if err != nil {
    base.Error().Printf("Failed to create RosterChooser: %v", err)
    return nil
  }
  rc.Request_dims = gui.Dims{
    rc.layout.Down.Data().Dx() + rc.layout.Option.Dx,
    rc.layout.Num_options * rc.layout.Option.Dy + 2*int(base.GetDictionary(15).MaxHeight()),
  }
  rc.selected = make(map[int]bool)
  rc.selection_filter = filter
  rc.on_complete = on_complete
  rc.render.options = make([]gui.Region, len(rc.options))
  return &rc
}

func (rc *RosterChooser) Think(ui *gui.Gui, t int64) {
  var dt int64
  if rc.last_think != 0 {
    dt = t - rc.last_think
  }
  rc.last_think = t
  for i := range rc.options {
    rc.options[i].Think(dt)
  }

  max := len(rc.options) - rc.layout.Num_options
  if rc.focus > max {
    rc.focus = max
  }
  if rc.focus < 0 {
    rc.focus = 0
  }
  rc.focus_pos = (1-rc.layout.Speed) * rc.focus_pos + rc.layout.Speed * float64(rc.focus)

  rc.mouse.X, rc.mouse.Y = gin.In().GetCursor("Mouse").Point()
}

func (rc *RosterChooser) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if found, event := group.FindEvent('l'); found && event.Type == gin.Press {
    rc.focus+=rc.layout.Num_options
    return true
  }
  if found, event := group.FindEvent('o'); found && event.Type == gin.Press {
    rc.focus-=rc.layout.Num_options
    return true
  }
  if found, event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    x, y := event.Key.Cursor().Point()
    gp := gui.Point{x, y}
    if gp.Inside(rc.render.down) {
      rc.focus+=rc.layout.Num_options
      return true
    } else if gp.Inside(rc.render.up) {
      rc.focus-=rc.layout.Num_options
      return true
    } else if gp.Inside(rc.render.all_options) {
      for i := range rc.render.options {
        if gp.Inside(rc.render.options[i]) {
          if rc.selection_filter(i, rc.selected) {
            if rc.selected[i] {
              delete(rc.selected, i)
            } else {
              rc.selected[i] = true
            }
          }
          return true
        }
      }
    } else if gp.Inside(rc.render.done) {
      if rc.selection_filter(-1, rc.selected) {
        rc.on_complete(rc.selected)
      }
      return true
    }
  }
  return false
}

func (rc *RosterChooser) Draw(r gui.Region) {
  rc.Render_region = r
  r.PushClipPlanes()
  defer r.PopClipPlanes()
  gl.Enable(gl.TEXTURE_2D)

  {  // Up button
    x := r.X
    y := r.Y + r.Dy - rc.layout.Up.Data().Dy()
    rc.render.up.X = x
    rc.render.up.Y = y
    rc.render.up.Dx = rc.layout.Up.Data().Dx()
    rc.render.up.Dy = rc.layout.Up.Data().Dy()
    if rc.mouse.Inside(rc.render.up) {
      gl.Color4d(1, 1, 1, 1)
    } else {
      gl.Color4d(0.8, 0.8, 0.8, 1)
    }
    rc.layout.Up.Data().RenderNatural(x, y)
  }

  {  // Down button
    x := r.X
    y := r.Y + rc.layout.Down.Data().Dy()
    rc.render.down.X = x
    rc.render.down.Y = y
    rc.render.down.Dx = rc.layout.Down.Data().Dx()
    rc.render.down.Dy = rc.layout.Down.Data().Dy()
    if rc.mouse.Inside(rc.render.down) {
      gl.Color4d(1, 1, 1, 1)
    } else {
      gl.Color4d(0.8, 0.8, 0.8, 1)
    }
    rc.layout.Down.Data().RenderNatural(x, y)
  }

  {  // Options
    rc.render.all_options.X = r.X + rc.layout.Down.Data().Dx()
    rc.render.all_options.Y = r.Y + r.Dy - rc.layout.Num_options * rc.layout.Option.Dy
    rc.render.all_options.Dx = rc.layout.Option.Dx
    rc.render.all_options.Dy = rc.layout.Num_options * rc.layout.Option.Dy
    rc.render.all_options.PushClipPlanes()
    x := rc.render.all_options.X
    y := r.Y + r.Dy - rc.layout.Option.Dy + int(float64(rc.layout.Option.Dy) * rc.focus_pos)
    for i := range rc.options {
        rc.render.options[i] = gui.Region{
        gui.Point{x, y},
        gui.Dims{rc.layout.Option.Dx, rc.layout.Option.Dy},
      }
      hovered := rc.mouse.Inside(rc.render.options[i])
      selected := rc.selected[i]
      selectable := rc.selection_filter(i, rc.selected)
      rc.options[i].Draw(hovered, selected, selectable, rc.render.options[i])
      y-=rc.layout.Option.Dy
    }

    rc.render.all_options.PopClipPlanes()
  }

  {  // Text
    d := base.GetDictionary(15)
    x := r.X
    y := float64(r.Y) + d.MaxHeight() / 2
    x1 := float64(x + r.Dx / 3)
    x2 := float64(x + (2 * r.Dx) / 3)

    rc.render.done = gui.Region{
      gui.Point{x, r.Y},
      gui.Dims{r.Dx/2, int(d.MaxHeight()*2)},
    }
    rc.render.sure = gui.Region{
      gui.Point{x + r.Dx/2, r.Y},
      gui.Dims{r.Dx/2, int(d.MaxHeight()*2)},
    }

    if rc.mouse.Inside(rc.render.done) {
      gl.Color4d(1, 1, 1, 1)
    } else {
      gl.Color4d(0.6, 0.6, 0.6, 1)
    }
    d.RenderString("Done", x1, y, 0, d.MaxHeight(), gui.Center)

    if rc.mouse.Inside(rc.render.sure) {
      gl.Color4d(1, 1, 1, 1)
    } else {
      gl.Color4d(0.6, 0.6, 0.6, 1)
    }
    d.RenderString("Rawr", x2, y, 0, d.MaxHeight(), gui.Center)

  }
}

func (rc *RosterChooser) DrawFocused(gui.Region) {

}

func (rc *RosterChooser) String() string {
  return "roster chooser"
}

