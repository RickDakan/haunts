package game

import (
  "fmt"
  "path/filepath"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/opengl/gl"
)

type Button struct {
  X,Y int
  Texture texture.Object

  // Color - brighter when the mouse is over it
  shade float64

  // Function to run whenever the button is clicked
  f func(interface{})
}

// If x,y is inside the button's region then it will run its function and
// return true, otherwise it does nothing and returns false.
func (b *Button) handleClick(x,y int, data interface{}) bool {
  d := b.Texture.Data()
  if x < b.X || y < b.Y || x >= b.X + d.Dx || y >= b.Y + d.Dy {
    return false
  }
  b.f(data)
  return true
}

func (b *Button) RenderAt(x,y,mx,my int) {
  b.Texture.Data().Bind()
  tdx :=  + b.Texture.Data().Dx
  tdy :=  + b.Texture.Data().Dy
  if mx >= x + b.X && mx < x + b.X + tdx && my >= y + b.Y && my < y + b.Y + tdy {
    b.shade = b.shade * 0.9 + 0.1
  } else {
    b.shade = b.shade * 0.9 + 0.04
  }
  gl.Color4d(1, 1, 1, b.shade)
  gl.Begin(gl.QUADS)
    gl.TexCoord2d(0, 0)
    gl.Vertex2i(x + b.X, y + b.Y)

    gl.TexCoord2d(0, -1)
    gl.Vertex2i(x + b.X, y + b.Y + tdy)

    gl.TexCoord2d(1,-1)
    gl.Vertex2i(x + b.X + tdx, y + b.Y + tdy)

    gl.TexCoord2d(1, 0)
    gl.Vertex2i(x + b.X + tdx, y + b.Y)
  gl.End()
}

type Center struct {
  X,Y int
}

type TextArea struct {
  X,Y           int
  Size          int
  Justification string
}

func (t *TextArea) RenderString(s string) {
  var just gui.Justification
  switch t.Justification {
  case "center":
    just = gui.Center
  case "left":
    just = gui.Left
  case "right":
    just = gui.Right
  default:
    base.Warn().Printf("Unknown justification '%s' in main gui bar.", t.Justification)
    t.Justification = "center"
  }
  px := float64(t.X)
  py := float64(t.Y)
  d := base.GetDictionary(t.Size)
  d.RenderString(s, px, py, 0, d.MaxHeight(), just)
}

type MainBarLayout struct {
  EndTurn     Button
  UnitLeft    Button
  UnitRight   Button
  ActionLeft  Button
  ActionRight Button

  CenterStillFrame Center

  Background texture.Object
  Divider    texture.Object
  Name   TextArea
  Ap     TextArea
  Hp     TextArea
  Corpus TextArea
  Ego    TextArea

  Conditions struct {
    X,Y,Height,Width,Size,Spacing float64
  }

  Actions struct {
    X,Y,Width,Icon_size float64
    Count int
    Empty texture.Object
  }
}

type mainBarState struct {
  Actions struct {
    // target is the action that should be displayed as left-most,
    // pos is the action that is currently left-most, which can be fractional.
    scroll_target float64
    scroll_pos    float64

    selected Action

    // The clicked action was clicked in the Ui but hasn't been set as the
    // selected action yet because we can't SetCurrentAction during event
    // handling.
    clicked Action

    space float64
  }
  Conditions struct {
    scroll_pos float64
  }
  MouseOver struct {
    active   bool
    text     string
    location mouseOverLocation
  }
}

type mouseOverLocation int
const (
  mouseOverActions = iota
  mouseOverConditions
  mouseOverGear
)

type MainBar struct {
  layout MainBarLayout
  state  mainBarState
  region gui.Region

  // List of all buttons, just to make it easy to iterate through them.
  buttons []*Button

  ent *Entity

  game *Game

  // Position of the mouse
  mx,my int
}

func buttonFuncEndTurn(mbi interface{}) {
  mb := mbi.(*MainBar)
  mb.game.OnRound()
}
func buttonFuncActionLeft(mbi interface{}) {
  mb := mbi.(*MainBar)
  mb.state.Actions.scroll_target -= float64(mb.layout.Actions.Count)
}
func buttonFuncActionRight(mbi interface{}) {
  mb := mbi.(*MainBar)
  mb.state.Actions.scroll_target += float64(mb.layout.Actions.Count)
}
func buttonFuncUnitLeft(mbi interface{}) {
  mb := mbi.(*MainBar)
  mb.game.SetCurrentAction(nil)
  start_index := len(mb.game.Ents) - 1
  for i := 0; i < len(mb.game.Ents); i++ {
    if mb.game.Ents[i] == mb.ent {
      start_index = i
      break
    }
  }
  for i := start_index - 1; i >= 0; i-- {
    if mb.game.Ents[i].Side() == mb.game.Side {
      mb.game.selected_ent = mb.game.Ents[i]
      return
    }
  }
  for i := len(mb.game.Ents) - 1; i >= start_index; i-- {
    if mb.game.Ents[i].Side() == mb.game.Side {
      mb.game.selected_ent = mb.game.Ents[i]
      return
    }
  }
}
func buttonFuncUnitRight(mbi interface{}) {
  mb := mbi.(*MainBar)
  mb.game.SetCurrentAction(nil)
  start_index := 0
  for i := 0; i < len(mb.game.Ents); i++ {
    if mb.game.Ents[i] == mb.ent {
      start_index = i
      break
    }
  }
  for i := start_index + 1; i < len(mb.game.Ents); i++ {
    if mb.game.Ents[i].Side() == mb.game.Side {
      mb.game.selected_ent = mb.game.Ents[i]
      return
    }
  }
  for i := 0; i <= start_index; i++ {
    if mb.game.Ents[i].Side() == mb.game.Side {
      mb.game.selected_ent = mb.game.Ents[i]
      return
    }
  }
}

func MakeMainBar(game *Game) (*MainBar, error) {
  var mb MainBar
  datadir := base.GetDataDir()
  err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "main_bar.json"), "json", &mb.layout)
  if err != nil {
    return nil, err
  }
  mb.buttons = []*Button{
    &mb.layout.EndTurn,
    &mb.layout.UnitLeft,
    &mb.layout.UnitRight,
    &mb.layout.ActionLeft,
    &mb.layout.ActionRight,
  }
  mb.layout.EndTurn.f = buttonFuncEndTurn
  mb.layout.UnitLeft.f = buttonFuncUnitLeft
  mb.layout.UnitRight.f = buttonFuncUnitRight
  mb.layout.ActionLeft.f = buttonFuncActionLeft
  mb.layout.ActionRight.f = buttonFuncActionRight
  mb.game = game
  return &mb, nil
}
func (m *MainBar) Requested() gui.Dims {
  return gui.Dims{
    Dx: m.layout.Background.Data().Dx,
    Dy: m.layout.Background.Data().Dy,
  }
}

func (mb *MainBar) SelectEnt(ent *Entity) {
  if ent == mb.ent {
    return
  }
  mb.ent = ent
  mb.state = mainBarState{}

  if mb.ent == nil {
    return
  }
}

func (m *MainBar) Expandable() (bool, bool) {
  return false, false
}

func (m *MainBar) Rendered() gui.Region {
  return m.region
}

func pointInsideRect(px, py, x, y, dx, dy int) bool {
  return px >= x && py >= y && px < x + dx && py < y + dy
}

func (m *MainBar) Think(g *gui.Gui, t int64) {
  if m.ent != nil {
    // If an action is selected and we can't see it then we scroll just enough
    // so that we can.
    min := 0.0
    max := float64(len(m.ent.Actions) - m.layout.Actions.Count)
    selected_index := -1
    for i := range m.ent.Actions {
      if m.ent.Actions[i] == m.state.Actions.selected {
        selected_index = i
        break
      }
    }
    if selected_index != -1 {
      if min < float64(selected_index - m.layout.Actions.Count + 1) {
        min = float64(selected_index - m.layout.Actions.Count + 1)
      }
      if max > float64(selected_index) {
        max = float64(selected_index)
      }
    }
    m.state.Actions.selected = m.game.current_action
    if m.state.Actions.scroll_target > max {
      m.state.Actions.scroll_target = max
    }
    if m.state.Actions.scroll_target < min {
      m.state.Actions.scroll_target = min
    }

    if m.state.Actions.clicked != nil {
      if m.state.Actions.selected != m.state.Actions.clicked {
        if m.state.Actions.clicked.Preppable(m.ent, m.game) {
          m.state.Actions.clicked.Prep(m.ent, m.game)
          m.game.SetCurrentAction(m.state.Actions.clicked)
        }
      }
      m.state.Actions.clicked = nil
    }

    // We similarly need to scroll through conditions
    c := m.layout.Conditions
    d := base.GetDictionary(int(c.Size))
    max_scroll := d.MaxHeight() * float64(len(m.ent.Stats.ConditionNames()))
    max_scroll -= m.layout.Conditions.Height
    // This might end up with a max that is negative, but we'll cap it at zero
    if m.state.Conditions.scroll_pos > max_scroll {
      m.state.Conditions.scroll_pos = max_scroll
    }
    if m.state.Conditions.scroll_pos < 0 {
      m.state.Conditions.scroll_pos = 0
    }
  } else {
    m.state.Conditions.scroll_pos = 0
    m.state.Actions.scroll_pos = 0
    m.state.Actions.scroll_target = 0
  }

  // Do a nice scroll motion towards the target position
  m.state.Actions.scroll_pos *= 0.8
  m.state.Actions.scroll_pos += 0.2 * m.state.Actions.scroll_target



  // Handle mouseover stuff after doing all of the scroll stuff since we don't
  // want to give a mouseover for something that the mouse isn't over after
  // scrolling something.
  m.state.MouseOver.active = false
  c := m.layout.Conditions
  if pointInsideRect(m.mx, m.my, int(c.X), int(c.Y), int(c.Width), int(c.Height)) {
    if m.ent == nil {
      m.state.MouseOver.active = false
    } else {
      pos := c.Y + c.Height + m.state.Conditions.scroll_pos - float64(m.my)
      index := int(pos / base.GetDictionary(int(c.Size)).MaxHeight())
      if index >= 0 && index < len(m.ent.Stats.ConditionNames()) {
        m.state.MouseOver.active = true
        m.state.MouseOver.text = m.ent.Stats.ConditionNames()[index]
        m.state.MouseOver.location = mouseOverConditions
      }
    }
  }
}

func (m *MainBar) Respond(g *gui.Gui, group gui.EventGroup) bool {
  cursor := group.Events[0].Key.Cursor()
  if cursor != nil {
    m.mx, m.my = cursor.Point()
  }

  if found, event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    for _, button := range m.buttons {
      if button.handleClick(m.mx, m.my, m) {
        return true
      }
    }
    if m.ent != nil {
      x := int(m.layout.Actions.X)
      y := int(m.layout.Actions.Y)
      x2 := int(m.layout.Actions.X + m.layout.Actions.Width)
      y2 := int(m.layout.Actions.Y + m.layout.Actions.Icon_size)
      if m.mx >= x && m.my >= y && m.mx < x2 && m.my < y2 {
        pos := float64(m.mx - x) / (m.layout.Actions.Icon_size + m.state.Actions.space)
        pos += m.state.Actions.scroll_pos
        p := int(pos)
        frac := pos - float64(p)
        // Make sure that the click didn't land in the space between two icons
        if frac < m.layout.Actions.Icon_size / (m.layout.Actions.Icon_size + m.state.Actions.space) {
          if p >= 0 && p < len(m.ent.Actions) {
            m.state.Actions.clicked = m.ent.Actions[p]
          }
        }
      }
    }
  }

  if found, event := group.FindEvent(gin.MouseWheelVertical); found {
    x := int(m.layout.Conditions.X)
    y := int(m.layout.Conditions.Y)
    x2 := int(m.layout.Conditions.X + m.layout.Conditions.Width)
    y2 := int(m.layout.Conditions.Y + m.layout.Conditions.Height)
    if m.mx >= x && m.my >= y && m.mx < x2 && m.my < y2 {
      m.state.Conditions.scroll_pos += event.Key.FramePressAmt()
    }
  }

  return false
}

func (m *MainBar) Draw(region gui.Region) {
  m.region = region
  gl.Enable(gl.TEXTURE_2D)
  m.layout.Background.Data().Bind()
  gl.Color4d(1, 1, 1, 1)
  gl.Begin(gl.QUADS)
    gl.TexCoord2d(0, 0)
    gl.Vertex2i(region.X, region.Y)

    gl.TexCoord2d(0, -1)
    gl.Vertex2i(region.X, region.Y + region.Dy)

    gl.TexCoord2d(1,-1)
    gl.Vertex2i(region.X + region.Dx, region.Y + region.Dy)

    gl.TexCoord2d(1, 0)
    gl.Vertex2i(region.X + region.Dx, region.Y)
  gl.End()

  for _, button := range m.buttons {
    button.RenderAt(region.X, region.Y, m.mx, m.my)
  }

  if m.ent != nil {
    gl.Color4d(1, 1, 1, 1)
    m.ent.Still.Data().Bind()
    tdx := m.ent.Still.Data().Dx
    tdy := m.ent.Still.Data().Dy
    cx := region.X + m.layout.CenterStillFrame.X
    cy := region.Y + m.layout.CenterStillFrame.Y
    gl.Begin(gl.QUADS)
      gl.TexCoord2d(0, 0)
      gl.Vertex2i(cx - tdx / 2, cy - tdy / 2)

      gl.TexCoord2d(0, -1)
      gl.Vertex2i(cx - tdx / 2, cy + tdy / 2)

      gl.TexCoord2d(1,-1)
      gl.Vertex2i(cx + tdx / 2, cy + tdy / 2)

      gl.TexCoord2d(1, 0)
      gl.Vertex2i(cx + tdx / 2, cy - tdy / 2)
    gl.End()

    m.layout.Name.RenderString(m.ent.Name)
    m.layout.Ap.RenderString(fmt.Sprintf("Ap:%d", m.ent.Stats.ApCur()))
    m.layout.Hp.RenderString(fmt.Sprintf("Hp:%d", m.ent.Stats.HpCur()))
    m.layout.Corpus.RenderString(fmt.Sprintf("Corpus:%d", m.ent.Stats.Corpus()))
    m.layout.Ego.RenderString(fmt.Sprintf("Ego:%d", m.ent.Stats.Ego()))

    gl.Color4d(1, 1, 1, 1)
    m.layout.Divider.Data().Bind()
    tdx = m.layout.Divider.Data().Dx
    tdy = m.layout.Divider.Data().Dy
    cx = region.X + m.layout.Name.X
    cy = region.Y + m.layout.Name.Y - 5
    gl.Begin(gl.QUADS)
      gl.TexCoord2d(0, 0)
      gl.Vertex2i(cx - tdx / 2, cy - tdy / 2)

      gl.TexCoord2d(0, -1)
      gl.Vertex2i(cx - tdx / 2, cy + (tdy + 1) / 2)

      gl.TexCoord2d(1,-1)
      gl.Vertex2i(cx + (tdx + 1) / 2, cy + (tdy + 1) / 2)

      gl.TexCoord2d(1, 0)
      gl.Vertex2i(cx + (tdx + 1) / 2, cy - tdy / 2)
    gl.End()

    // Actions
    {
      spacing := m.layout.Actions.Icon_size * float64(m.layout.Actions.Count)
      spacing = m.layout.Actions.Width - spacing
      spacing /= float64(m.layout.Actions.Count - 1)
      m.state.Actions.space = spacing
      s := m.layout.Actions.Icon_size
      num_actions := len(m.ent.Actions)
      xpos := m.layout.Actions.X

      if num_actions > m.layout.Actions.Count {
        xpos -= m.state.Actions.scroll_pos * (s + spacing)
      }
      d := base.GetDictionary(10)
      var r gui.Region
      r.X = int(m.layout.Actions.X)
      r.Y = int(m.layout.Actions.Y - d.MaxHeight())
      r.Dx = int(m.layout.Actions.Width)
      r.Dy = int(m.layout.Actions.Icon_size + d.MaxHeight())
      r.PushClipPlanes()

      gl.Color4d(1, 1, 1, 1)
      for i,action := range m.ent.Actions {

        // Highlight the selected action
        if action == m.game.current_action {
          gl.Disable(gl.TEXTURE_2D)
          gl.Color4d(1, 0, 0, 1)
          gl.Begin(gl.QUADS)
            gl.Vertex3d(xpos - 2, m.layout.Actions.Y - 2, 0)
            gl.Vertex3d(xpos - 2, m.layout.Actions.Y+s + 2, 0)
            gl.Vertex3d(xpos+s + 2, m.layout.Actions.Y+s + 2, 0)
            gl.Vertex3d(xpos+s + 2, m.layout.Actions.Y - 2, 0)
          gl.End()
        }
        gl.Enable(gl.TEXTURE_2D)
        action.Icon().Data().Bind()
        if action.Preppable(m.ent, m.game) {
          gl.Color4d(1, 1, 1, 1)
        } else {
          gl.Color4d(0.5, 0.5, 0.5, 1)
        }
        gl.Begin(gl.QUADS)
          gl.TexCoord2d(0, 0)
          gl.Vertex3d(xpos, m.layout.Actions.Y, 0)

          gl.TexCoord2d(0, -1)
          gl.Vertex3d(xpos, m.layout.Actions.Y+s, 0)

          gl.TexCoord2d(1, -1)
          gl.Vertex3d(xpos+s, m.layout.Actions.Y+s, 0)

          gl.TexCoord2d(1, 0)
          gl.Vertex3d(xpos+s, m.layout.Actions.Y, 0)
        gl.End()
        gl.Disable(gl.TEXTURE_2D)

        ypos := m.layout.Actions.Y - d.MaxHeight() - 2
        d.RenderString(fmt.Sprintf("%d", i+1), xpos + s / 2, ypos, 0, d.MaxHeight(), gui.Center)

        xpos += spacing + m.layout.Actions.Icon_size
      }

      r.PopClipPlanes()

      // Now, if there is a selected action, position it between the arrows
      if m.state.Actions.selected != nil {
        // a := m.state.Actions.selected
        d := base.GetDictionary(15)
        x := m.layout.Actions.X + m.layout.Actions.Width / 2
        y := float64(m.layout.ActionLeft.Y)
        str := fmt.Sprintf("%s:%dAP", m.state.Actions.selected.String(), m.state.Actions.selected.AP())
        gl.Color4d(1, 1, 1, 1)
        d.RenderString(str, x, y, 0, d.MaxHeight(), gui.Center)
      }
    }

    // Conditions
    {
      gl.Color4d(1, 1, 1, 1)
      c := m.layout.Conditions
      d := base.GetDictionary(int(c.Size))
      ypos := c.Y + c.Height - d.MaxHeight() + m.state.Conditions.scroll_pos
      var r gui.Region
      r.X = int(c.X)
      r.Y = int(c.Y)
      r.Dx = int(c.Width)
      r.Dy = int(c.Height)
      r.PushClipPlanes()
      for _,s := range m.ent.Stats.ConditionNames() {
        d.RenderString(s, c.X + c.Width / 2, ypos, 0, d.MaxHeight(), gui.Center)
        ypos -= float64(d.MaxHeight())
      }

      r.PopClipPlanes()
    }
  }

  // Mouseover text
  if m.state.MouseOver.active {
    var x int
    switch m.state.MouseOver.location {
      case mouseOverActions:
        x = int(m.layout.Actions.X + m.layout.Actions.Width / 2)
      case mouseOverConditions:
        x = int(m.layout.Conditions.X + m.layout.Conditions.Width / 2)
      case mouseOverGear:
      default:
        base.Warn().Printf("Got an unknown mouseover location: %d", m.state.MouseOver.location)
        m.state.MouseOver.active = false
    }
    y := m.layout.Background.Data().Dy - 40
    d := base.GetDictionary(15)
    d.RenderString(m.state.MouseOver.text, float64(x), float64(y), 0, d.MaxHeight(), gui.Center)
  }
}

func (m *MainBar) DrawFocused(region gui.Region) {

}

func (m *MainBar) String() string {
  return "main bar"
}

