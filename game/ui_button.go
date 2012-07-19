package game

import (
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/opengl/gl"
)

type Button struct {
  X, Y    int
  Texture texture.Object
  Text    struct {
    String        string
    Size          int
    Justification string
  }

  // Color - brighter when the mouse is over it
  shade float64

  // Function to run whenever the button is clicked
  f func(interface{})

  // Key that can be bound to have the same effect as clicking this button
  key gin.KeyId

  bounds struct {
    x, y, dx, dy int
  }
}

// If x,y is inside the button's region then it will run its function and
// return true, otherwise it does nothing and returns false.
func (b *Button) handleClick(x, y int, data interface{}) bool {
  in := pointInsideRect(x, y, b.bounds.x, b.bounds.y, b.bounds.dx, b.bounds.dy)
  if in {
    b.f(data)
  }
  base.Log().Printf("Dims: %d %d %d %d %d %d", x, y, b.bounds.x, b.bounds.y, b.bounds.dx, b.bounds.dy)
  base.Log().Printf("Point inside: %t", in)
  return in
}

func (b *Button) Respond(group gui.EventGroup, data interface{}) bool {
  if group.Events[0].Key.Id() == b.key && group.Events[0].Type == gin.Press {
    b.f(data)
    return true
  }
  return false
}

func doShading(current float64, in bool, dt int64) float64 {
  if in {
    return current*0.9 + 0.1
  }
  if current < 0.40 {
    return 0.40
  }
  return current*0.9 + 0.04
}

func (b *Button) Think(x, y, mx, my int, dt int64) {
  // var tdx, tdy int
  // if b.Texture.Data() != nil {
  //   tdx = b.Texture.Data().Dx()
  //   tdy = b.Texture.Data().Dy()
  // } else {
  // }
  in := pointInsideRect(mx, my, b.bounds.x, b.bounds.y, b.bounds.dx, b.bounds.dy)
  b.shade = doShading(b.shade, in, dt)
}

func (b *Button) RenderAt(x, y int) {
  gl.Color4ub(255, 255, 255, byte(b.shade*255))
  if b.Texture.Path != "" {
    b.Texture.Data().RenderNatural(b.X+x, b.Y+y)
    b.bounds.x = b.X + x
    b.bounds.y = b.Y + y
    b.bounds.dx = b.Texture.Data().Dx()
    b.bounds.dy = b.Texture.Data().Dy()
  } else {
    d := base.GetDictionary(b.Text.Size)
    b.bounds.x = b.X + x
    b.bounds.y = b.Y + y
    b.bounds.dx = int(d.StringWidth(b.Text.String))
    b.bounds.dy = int(d.MaxHeight())
    var just gui.Justification
    switch b.Text.Justification {
    case "center":
      just = gui.Center
      b.bounds.x -= b.bounds.dx / 2
    case "left":
      just = gui.Left
    case "right":
      just = gui.Right
      b.bounds.x -= b.bounds.dx
    default:
      just = gui.Center
      b.bounds.x -= b.bounds.dx / 2
      b.Text.Justification = "center"
      base.Warn().Printf("Failed to indicate valid aligmnent, '%s' is not valid.", b.Text.Justification)
    }
    d.RenderString(b.Text.String, float64(b.X+x), float64(b.Y+y), 0, d.MaxHeight(), just)
  }
}
