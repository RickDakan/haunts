package house

import (
  "hash/fnv"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/mathgl"
  gl "github.com/chsc/gogl/gl21"
)

type SpawnPoint struct {
  Name  string
  Dx,Dy int
  X,Y   int

  // just for the shader
  temporary, invalid bool
}
func (sp *SpawnPoint) Dims() (int,int) {
  return sp.Dx, sp.Dy
}
func (sp *SpawnPoint) Pos() (int,int) {
  return sp.X, sp.Y
}
func (sp *SpawnPoint) FPos() (float64,float64) {
  return float64(sp.X), float64(sp.Y)
}
func (sp *SpawnPoint) Color() (r,g,b,a byte) {
  return 255, 255, 255, 255
}
func (sp *SpawnPoint) Render(pos mathgl.Vec2, width float32) {
  gl.Disable(gl.TEXTURE_2D)
  gl.Color4d(1, 1, 1, 0.1)
  gl.Begin(gl.QUADS)
    gl.Vertex2f(pos.X-width/2, pos.Y)
    gl.Vertex2f(pos.X-width/2, pos.Y+width)
    gl.Vertex2f(pos.X+width/2, pos.Y+width)
    gl.Vertex2f(pos.X+width/2, pos.Y)
  gl.End()
}
func (sp *SpawnPoint) RenderOnFloor() {
  var rgba [4]float64
  gl.GetDoublev(gl.CURRENT_COLOR, &rgba[0])
  gl.PushAttrib(gl.CURRENT_BIT)
  gl.Disable(gl.TEXTURE_2D)

  prefix := sp.Name
  for i := range prefix {
    if prefix[i] == '-' {
      prefix = prefix[0:i]
      break
    }
  }
  h := fnv.New32()
  h.Write([]byte(prefix))
  hs := h.Sum32()
  gl.Color4ub(byte(hs % 256), byte((hs / 256) % 256), byte((hs / (256*256)) % 256), byte(255 * rgba[3]))

  base.EnableShader("box")
  base.SetUniformF("box", "dx", float32(sp.Dx))
  base.SetUniformF("box", "dy", float32(sp.Dy))
  if !sp.temporary {
    base.SetUniformI("box", "temp_invalid", 0)
  } else if !sp.invalid {
    base.SetUniformI("box", "temp_invalid", 1)
  } else {
    base.SetUniformI("box", "temp_invalid", 2)
  }
  (&texture.Object{}).Data().Render(float64(sp.X), float64(sp.Y), float64(sp.Dx), float64(sp.Dy))
  base.EnableShader("")
  gl.PopAttrib()
}
