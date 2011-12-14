package house

import (
  "fmt"
  "glop/gui"
  "glop/gin"
  "glop/sprite"
  "os"
  "image"
  "image/draw"
  _ "image/png"
  _ "image/jpeg"
  "gl"
  "gl/glu"
  "math"
  "github.com/arbaal/mathgl"
)

func init() {
  fmt.Printf("")
}

type RoomViewer struct {
  gui.Childless
  gui.EmbeddedWidget
  gui.BasicZone
  gui.NonThinker
  gui.NonFocuser

  // Length of the side of block in the source image.
  block_size int

  // All events received by the viewer are passed to the handler
  handler gin.EventHandler

  // Focus, in map coordinates
  fx, fy float32

  // The viewing angle, 0 means the map is viewed head-on, 90 means the map is viewed
  // on its edge (i.e. it would not be visible)
  angle float32

  // Zoom factor, 1.0 is standard
  zoom float32

  // The modelview matrix that is sent to opengl.  Updated any time focus, zoom, or viewing
  // angle changes
  mat mathgl.Mat4

  // Inverse of mat
  imat mathgl.Mat4

  // All drawables that will be drawn parallel to the window
  upright_drawables []sprite.ZDrawable
  upright_positions []mathgl.Vec3

  // All drawables that will be drawn on the surface of the viewer
  flattened_drawables []sprite.ZDrawable
  flattened_positions []mathgl.Vec3

  // Don't need to keep the image around once it's loaded into texture memory,
  // only need to keep around the dimensions
  bg_dims gui.Dims
  texture gl.Texture
}

func (rv *RoomViewer) String() string {
  return "viewer"
}

func (rv *RoomViewer) AddUprightDrawable(x, y float32, zd sprite.ZDrawable) {
  rv.upright_drawables = append(rv.upright_drawables, zd)
  rv.upright_positions = append(rv.upright_positions, mathgl.Vec3{x, y, 0})
}

// x,y: board coordinates that the drawable should be drawn arv.
// zd: drawable that will be rendered after the viewer has been rendered, it will be rendered
//     with the same modelview matrix as the rest of the viewer
func (rv *RoomViewer) AddFlattenedDrawable(x, y float32, zd sprite.ZDrawable) {
  rv.flattened_drawables = append(rv.flattened_drawables, zd)
  rv.flattened_positions = append(rv.flattened_positions, mathgl.Vec3{x, y, 0})
}

func MakeRoomViewer(bg_path string, block_size, dx, dy int, angle float32) (*RoomViewer, error) {
  var rv RoomViewer
  rv.EmbeddedWidget = &gui.BasicWidget{CoreWidget: &rv}

  f, err := os.Open(bg_path)
  if err != nil {
    return nil, err
  }
  defer f.Close()
  bg, _, err := image.Decode(f)
  if err != nil {
    return nil, err
  }

  rv.block_size = block_size
  rv.angle = angle

  rv.bg_dims.Dx = bg.Bounds().Dx()
  rv.bg_dims.Dy = bg.Bounds().Dy()
  rgba := image.NewRGBA(image.Rect(0, 0, rv.bg_dims.Dx, rv.bg_dims.Dy))
  draw.Draw(rgba, bg.Bounds(), bg, image.Point{0, 0}, draw.Over)

  gl.Enable(gl.TEXTURE_2D)
  rv.texture = gl.GenTexture()
  rv.texture.Bind(gl.TEXTURE_2D)
  gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
  glu.Build2DMipmaps(gl.TEXTURE_2D, 4, rv.bg_dims.Dx, rv.bg_dims.Dy, gl.RGBA, rgba.Pix)

  rv.Zoom(-1)
  rv.fx = float32(rv.bg_dims.Dx / rv.block_size / 2)
  rv.fy = float32(rv.bg_dims.Dy / rv.block_size / 2)

  rv.makeMat()
  rv.Request_dims.Dx = 100
  rv.Request_dims.Dy = 100
  rv.Ex = true
  rv.Ey = true
  return &rv, nil
}

func (rv *RoomViewer) makeMat() {
  var m mathgl.Mat4
  rv.mat.Translation(float32(rv.Render_region.Dx/2+rv.Render_region.X), float32(rv.Render_region.Dy/2+rv.Render_region.Y), 0)
  m.RotationZ(45 * math.Pi / 180)
  rv.mat.Multiply(&m)
  m.RotationAxisAngle(mathgl.Vec3{X: -1, Y: 1}, -float32(rv.angle)*math.Pi/180)
  rv.mat.Multiply(&m)

  s := float32(rv.zoom)
  m.Scaling(s, s, s)
  rv.mat.Multiply(&m)

  // Move the viewer so that (rv.fx,rv.fy) is at the origin, and hence becomes centered
  // in the window
  xoff := (rv.fx + 0.5) * float32(rv.block_size)
  yoff := (rv.fy + 0.5) * float32(rv.block_size)
  m.Translation(-xoff, -yoff, 0)
  rv.mat.Multiply(&m)

  rv.imat.Assign(&rv.mat)
  rv.imat.Inverse()
}

// Transforms a cursor position in window coordinates to board coordinates.  Does not check
// to make sure that the values returned represent a valid position on the board.
func (rv *RoomViewer) WindowToBoard(wx, wy int) (float32, float32) {
  mx := float32(wx)
  my := float32(wy)
  return rv.modelviewToBoard(mx, my)
}

func (rv *RoomViewer) modelviewToBoard(mx, my float32) (float32, float32) {
  mz := (my - float32(rv.Render_region.Y+rv.Render_region.Dy/2)) * float32(math.Tan(float64(rv.angle*math.Pi/180)))
  v := mathgl.Vec4{X: mx, Y: my, Z: mz, W: 1}
  v.Transform(&rv.imat)
  return v.X / float32(rv.block_size), v.Y / float32(rv.block_size)
}

func (rv *RoomViewer) boardToModelview(mx, my float32) (x, y, z float32) {
  v := mathgl.Vec4{X: mx * float32(rv.block_size), Y: my * float32(rv.block_size), W: 1}
  v.Transform(&rv.mat)
  x, y, z = v.X, v.Y, v.Z
  return
}

func clamp(f, min, max float32) float32 {
  if f < min {
    return min
  }
  if f > max {
    return max
  }
  return f
}

// The change in x and y screen coordinates to apply to point on the viewer the is in
// focus.  These coordinates will be scaled by the current zoom.
func (rv *RoomViewer) Move(dx, dy float64) {
  if dx == 0 && dy == 0 {
    return
  }
  dy /= math.Sin(float64(rv.angle) * math.Pi / 180)
  dx, dy = dy+dx, dy-dx
  rv.fx += float32(dx) / rv.zoom
  rv.fy += float32(dy) / rv.zoom
  rv.fx = clamp(rv.fx, 0, float32(rv.bg_dims.Dx/rv.block_size))
  rv.fy = clamp(rv.fy, 0, float32(rv.bg_dims.Dy/rv.block_size))
  rv.makeMat()
}

// Changes the current zoom from e^(zoom) to e^(zoom+dz)
func (rv *RoomViewer) Zoom(dz float64) {
  if dz == 0 {
    return
  }
  exp := math.Log(float64(rv.zoom)) + dz
  exp = float64(clamp(float32(exp), -1.25, 1.25))
  rv.zoom = float32(math.Exp(exp))
  rv.makeMat()
}

func (rv *RoomViewer) Draw(region gui.Region) {
  region.PushClipPlanes()
  defer region.PopClipPlanes()
  if rv.Render_region.X != region.X || rv.Render_region.Y != region.Y || rv.Render_region.Dx != region.Dx || rv.Render_region.Dy != region.Dy {
    rv.Render_region = region
    rv.makeMat()
  }
  gl.MatrixMode(gl.MODELVIEW)
  gl.PushMatrix()
  gl.LoadIdentity()
  gl.MultMatrixf(&rv.mat[0])
  defer gl.PopMatrix()

  gl.Disable(gl.DEPTH_TEST)
  gl.Disable(gl.TEXTURE_2D)
  gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
  gl.Color3d(1, 0, 0)
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
  fdx := float32(rv.bg_dims.Dx)
  fdy := float32(rv.bg_dims.Dy)

  // Draw a simple border around the viewer
  gl.Color4d(1, .3, .3, 1)
  gl.Begin(gl.QUADS)
  fbs := float32(rv.block_size)
  gl.Vertex2f(-fbs, -fbs)
  gl.Vertex2f(-fbs, fdy+fbs)
  gl.Vertex2f(fdx+fbs, fdy+fbs)
  gl.Vertex2f(fdx+fbs, -fbs)
  gl.End()

  gl.Enable(gl.TEXTURE_2D)
  rv.texture.Bind(gl.TEXTURE_2D)
  gl.Color4d(1.0, 1.0, 1.0, 1.0)
  gl.Begin(gl.QUADS)
  gl.TexCoord2f(0, 0)
  gl.Vertex2f(0, 0)
  gl.TexCoord2f(0, -1)
  gl.Vertex2f(0, fdy)
  gl.TexCoord2f(1, -1)
  gl.Vertex2f(fdx, fdy)
  gl.TexCoord2f(1, 0)
  gl.Vertex2f(fdx, 0)
  gl.End()

  gl.Disable(gl.TEXTURE_2D)
  gl.Color4f(0, 0, 0, 0.5)
  gl.Begin(gl.LINES)
  for i := float32(0); i < float32(rv.bg_dims.Dx); i += float32(rv.block_size) {
    gl.Vertex2f(i, 0)
    gl.Vertex2f(i, float32(rv.bg_dims.Dy))
  }
  for j := float32(0); j < float32(rv.bg_dims.Dy); j += float32(rv.block_size) {
    gl.Vertex2f(0, j)
    gl.Vertex2f(float32(rv.bg_dims.Dx), j)
  }
  gl.End()

  for i := range rv.flattened_positions {
    v := rv.flattened_positions[i]
    rv.flattened_drawables[i].Render(v.X, v.Y, 0, float32(rv.block_size))
  }
  rv.flattened_positions = rv.flattened_positions[0:0]
  rv.flattened_drawables = rv.flattened_drawables[0:0]

  for i := range rv.upright_positions {
    vx, vy, vz := rv.boardToModelview(rv.upright_positions[i].X, rv.upright_positions[i].Y)
    rv.upright_positions[i] = mathgl.Vec3{vx, vy, vz}
  }
  sprite.ZSort(rv.upright_positions, rv.upright_drawables)
  gl.Disable(gl.TEXTURE_2D)
  gl.PushMatrix()
  gl.LoadIdentity()
  for i := range rv.upright_positions {
    v := rv.upright_positions[i]
    rv.upright_drawables[i].Render(v.X, v.Y, v.Z, float32(rv.zoom))
  }
  rv.upright_positions = rv.upright_positions[0:0]
  rv.upright_drawables = rv.upright_drawables[0:0]
  gl.PopMatrix()
}

func (rv *RoomViewer) SetEventHandler(handler gin.EventHandler) {
  rv.handler = handler
}

func (rv *RoomViewer) DoRespond(event_group gui.EventGroup) (bool, bool) {
  if rv.handler != nil {
    rv.handler.HandleEventGroup(event_group.EventGroup)
  }
  return false, false
}
