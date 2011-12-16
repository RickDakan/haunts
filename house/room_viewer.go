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
  "sort"
)

type RectObject interface {
  // Position in board coordinates
  Pos() mathgl.Vec2

  // Dimensions in board coordinates
  Dims() mathgl.Vec2

  Render(pos mathgl.Vec2, width float32)
}

type furniture struct {
  td textureData
  pos mathgl.Vec2
  dims mathgl.Vec2
  image_dims mathgl.Vec2
}
func (f *furniture) Pos() mathgl.Vec2 {
  return f.pos
}
func (f *furniture) Dims() mathgl.Vec2 {
  return f.dims
}
func (f *furniture) Render(pos mathgl.Vec2, width float32) {
  gl.Disable(gl.TEXTURE_2D)
  gl.Color4d(1, 0, 0, 0.3)
  // Draw a furniture tile
  dy := width * (f.image_dims.Y / f.image_dims.X)
  // gl.Begin(gl.QUADS)
  // gl.Vertex2f(pos.X, pos.Y)
  // gl.Vertex2f(pos.X, pos.Y + dy)
  // gl.Vertex2f(pos.X + width, pos.Y + dy)
  // gl.Vertex2f(pos.X + width, pos.Y)
  // gl.End()

  gl.Enable(gl.TEXTURE_2D)
  gl.Color4d(1, 1, 1, 1)
  // Draw a furniture tile
  f.td.texture.Bind(gl.TEXTURE_2D)
  gl.Begin(gl.QUADS)
  gl.TexCoord2f(0, 1)
  gl.Vertex2f(pos.X, pos.Y)
  gl.TexCoord2f(0, 0)
  gl.Vertex2f(pos.X, pos.Y + dy)
  gl.TexCoord2f(1, 0)
  gl.Vertex2f(pos.X + width, pos.Y + dy)
  gl.TexCoord2f(1, 1)
  gl.Vertex2f(pos.X + width, pos.Y)
  gl.End()
}




type rectObjectArray []RectObject
func (r rectObjectArray) Len() int {
  return len(r)
}
func (r rectObjectArray) Swap(i,j int) {
  r[i],r[j] = r[j],r[i]
}
func (r rectObjectArray) Less(i,j int) bool {
  idims := r[i].Dims()
  ifar := r[i].Pos()
  ifar.Add(&idims)

  jdims := r[j].Dims()
  jfar := r[j].Pos()
  jfar.Add(&jdims)

  if ifar.X <= r[j].Pos().X { return false }
  if ifar.Y <= r[j].Pos().Y { return false }
  if jfar.X <= r[i].Pos().X { return true }
  if jfar.Y <= r[i].Pos().Y { return true }

  if r[i].Pos().X != r[j].Pos().X {
    return r[i].Pos().X < r[j].Pos().X
  }
  return r[i].Pos().Y < r[j].Pos().Y
}


type RoomViewer struct {
  gui.Childless
  gui.EmbeddedWidget
  gui.BasicZone
  gui.NonThinker
  gui.NonFocuser

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

  furn rectObjectArray

  dx,dy int

  floor textureData
  floor_reload_output chan textureData
  floor_reload_input chan string

  wall textureData
  wall_reload_output chan textureData
  wall_reload_input chan string

  table textureData
  table_reload_output chan textureData
  table_reload_input chan string

  cube textureData
  cube_reload_output chan textureData
  cube_reload_input chan string

  // // Don't need to keep the image around once it's loaded into texture memory,
  // // only need to keep around the dimensions
  // bg_dims gui.Dims
  // texture gl.Texture
}

func (rv *RoomViewer) ReloadFloor(path string) {
  rv.floor_reload_input <- path
}

func (rv *RoomViewer) ReloadWall(path string) {
  rv.wall_reload_input <- path
}
func bindToTexture(td *textureData) {
  gl.Enable(gl.TEXTURE_2D)
  td.texture = gl.GenTexture()
  td.texture.Bind(gl.TEXTURE_2D)
  gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
  glu.Build2DMipmaps(gl.TEXTURE_2D, 4, td.dims.Dx, td.dims.Dy, gl.RGBA, td.image.Pix)
  td.image = nil
}
func (rv *RoomViewer) reloader() {
  done := false
  for !done {
    select {
    case td := <-rv.floor_reload_output:
      rv.floor.texture.Delete()
      rv.floor = td
      bindToTexture(&rv.floor)

    case td := <-rv.wall_reload_output:
      rv.wall.texture.Delete()
      rv.wall = td
      bindToTexture(&rv.wall)

    case td := <-rv.table_reload_output:
      fmt.Printf("table reload: %v\n", td.texture)
      rv.table.texture.Delete()
      rv.table = td
      bindToTexture(&rv.table)
      f := &furniture{
        td: rv.table,
        pos: mathgl.Vec2{1,3},
        dims: mathgl.Vec2{5,3},
        image_dims: mathgl.Vec2{ float32(td.image.Rect.Dx()), float32(td.image.Rect.Dy()) },
      }
      rv.furn = append(rv.furn, f)
      f = &furniture{
        td: rv.table,
        pos: mathgl.Vec2{9,7},
        dims: mathgl.Vec2{5,3},
        image_dims: mathgl.Vec2{ float32(td.image.Rect.Dx()), float32(td.image.Rect.Dy()) },
      }
      rv.furn = append(rv.furn, f)
      f = &furniture{
        td: rv.table,
        pos: mathgl.Vec2{3,6},
        dims: mathgl.Vec2{5,3},
        image_dims: mathgl.Vec2{ float32(td.image.Rect.Dx()), float32(td.image.Rect.Dy()) },
      }
      rv.furn = append(rv.furn, f)
      f = &furniture{
        td: rv.table,
        pos: mathgl.Vec2{7,2},
        dims: mathgl.Vec2{5,3},
        image_dims: mathgl.Vec2{ float32(td.image.Rect.Dx()), float32(td.image.Rect.Dy()) },
      }
      rv.furn = append(rv.furn, f)

    case td := <-rv.cube_reload_output:
      fmt.Printf("cube reload: %v\n", td.texture)
      rv.cube.texture.Delete()
      rv.cube = td
      bindToTexture(&rv.cube)
      f := &furniture{
        td: rv.cube,
        pos: mathgl.Vec2{2,2},
        dims: mathgl.Vec2{1,1},
        image_dims: mathgl.Vec2{ float32(td.image.Rect.Dx()), float32(td.image.Rect.Dy()) },
      }
      rv.furn = append(rv.furn, f)

    default:
      done = true
    }
  }
}

type textureData struct {
  dims    gui.Dims
  texture gl.Texture
  image   *image.RGBA
}

func textureDataReloadRoutine(input <-chan string, output chan<- textureData) {
  gathered := make(chan string)

  // This go-routine guarantees that there is no way someone can call Reload*()
  // several times in a row and dead-lock everything because they didn't call
  // Think() enough times while doing so.  This go-routine just collects
  // requests and queues them up to send them off later, essentially like a
  // buffered channel except of variable size.
  go func() {
    var pending []string
    var next,last string
    active := true
    for active {
      if len(pending) == 0 {
        next,active = <-input
        if next == last { continue }
        last = next
        if active {
          pending = append(pending, next)
        }
      } else {
        select {
        case next,active = <-input:
          if next == last { continue }
          last = next
          if active {
            pending = append(pending, next)
          }

        case gathered <- pending[0]:
          pending = pending[1:]
          if len(pending) == 0 {
            pending = nil
          }
        }
      }
    }
    for _,path := range pending {
      gathered <- path
    }
    close(gathered)
  } ()

  for path := range gathered {
    f, err := os.Open(path)
    if err != nil {
      fmt.Printf("Error: %v\n", err)
      // TODO: Log an error here
      continue
    }
    image_data, _, err := image.Decode(f)
    f.Close()
    if err != nil {
      fmt.Printf("Error: %v\n", err)
      // TODO: Log an error here
      continue
    }

    var td textureData

    td.dims.Dx = image_data.Bounds().Dx()
    td.dims.Dy = image_data.Bounds().Dy()

    td.image = image.NewRGBA(image.Rect(0, 0, td.dims.Dx, td.dims.Dy))
    draw.Draw(td.image, image_data.Bounds(), image_data, image.Point{0, 0}, draw.Over)

    output <- td
  }
  close(output)
}

func (rv *RoomViewer) String() string {
  return "viewer"
}

func (rv *RoomViewer) SetDims(dx,dy int) {
  rv.dx = dx
  rv.dy = dy
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

func MakeRoomViewer(dx, dy int, angle float32) *RoomViewer {
  var rv RoomViewer
  rv.EmbeddedWidget = &gui.BasicWidget{CoreWidget: &rv}

  rv.floor_reload_output = make(chan textureData)
  rv.floor_reload_input = make(chan string)
  rv.wall_reload_output = make(chan textureData)
  rv.wall_reload_input = make(chan string)
  rv.table_reload_output = make(chan textureData)
  rv.table_reload_input = make(chan string)
  rv.cube_reload_output = make(chan textureData)
  rv.cube_reload_input = make(chan string)

  go textureDataReloadRoutine(rv.floor_reload_input, rv.floor_reload_output)
  go textureDataReloadRoutine(rv.wall_reload_input, rv.wall_reload_output)
  go textureDataReloadRoutine(rv.table_reload_input, rv.table_reload_output)
  go textureDataReloadRoutine(rv.cube_reload_input, rv.cube_reload_output)
  rv.table_reload_input <- "/Users/runningwild/Downloads/table_02.png"
// /  rv.cube_reload_input <- "/Users/runningwild/Downloads/cube.png"
  rv.dx = dx
  rv.dy = dy
  rv.angle = angle
  rv.fx = float32(rv.dx / 2)
  rv.fy = float32(rv.dy / 2)
  rv.Zoom(1)
  rv.makeMat()
  rv.Request_dims.Dx = 100
  rv.Request_dims.Dy = 100
  rv.Ex = true
  rv.Ey = true



  return &rv
}

func (rv *RoomViewer) AdjAngle(ang float32) {
  rv.angle = ang
  rv.makeMat()
}

func (rv *RoomViewer) makeMat() {
  var m mathgl.Mat4
  rv.mat.Translation(float32(rv.Render_region.Dx/2+rv.Render_region.X), float32(rv.Render_region.Dy/2+rv.Render_region.Y), 0)

  // NOTE: If we want to change 45 to *anything* else then we need to do the
  // appropriate math for rendering quads for furniture
  m.RotationZ(45 * math.Pi / 180)
  rv.mat.Multiply(&m)
  m.RotationAxisAngle(mathgl.Vec3{X: -1, Y: 1}, -float32(rv.angle)*math.Pi/180)
  rv.mat.Multiply(&m)

  s := float32(rv.zoom)
  m.Scaling(s, s, s)
  rv.mat.Multiply(&m)

  // Move the viewer so that (rv.fx,rv.fy) is at the origin, and hence becomes centered
  // in the window
  xoff := rv.fx + 0.5
  yoff := rv.fy + 0.5
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
  return v.X, v.Y
}

func (rv *RoomViewer) boardToModelview(mx, my float32) (x, y, z float32) {
  v := mathgl.Vec4{X: mx, Y: my, W: 1}
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
  rv.fx = clamp(rv.fx, 0, float32(rv.dx))
  rv.fy = clamp(rv.fy, 0, float32(rv.dy))
  rv.makeMat()
}

// Changes the current zoom from e^(zoom) to e^(zoom+dz)
func (rv *RoomViewer) Zoom(dz float64) {
  if dz == 0 {
    return
  }
  exp := math.Log(float64(rv.zoom)) + dz
  exp = float64(clamp(float32(exp), 2.5, 5.0))
  rv.zoom = float32(math.Exp(exp))
  rv.makeMat()
}

func (rv *RoomViewer) Draw(region gui.Region) {
  rv.reloader()
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

  // Draw a simple border around the viewer
  gl.Color4d(0.1, 0.3, 0.8, 1)
  gl.Begin(gl.QUADS)
  gl.Vertex2i(-1, -1)
  gl.Vertex2i(-1, rv.dy+1)
  gl.Vertex2i(rv.dx+1, rv.dy+1)
  gl.Vertex2i(rv.dx+1, -1)
  gl.End()


  // Draw the floor
  gl.Enable(gl.TEXTURE_2D)
  rv.floor.texture.Bind(gl.TEXTURE_2D)
  gl.Color4d(1.0, 1.0, 1.0, 1.0)
  gl.Begin(gl.QUADS)
    gl.TexCoord2i(0, 0)
    gl.Vertex2i(0, 0)
    gl.TexCoord2i(0, -1)
    gl.Vertex2i(0, rv.dy)
    gl.TexCoord2i(1, -1)
    gl.Vertex2i(rv.dx, rv.dy)
    gl.TexCoord2i(1, 0)
    gl.Vertex2i(rv.dx, 0)
  gl.End()


  // Draw the wall
  rv.wall.texture.Bind(gl.TEXTURE_2D)
  corner := float32(rv.dx) / float32(rv.dx + rv.dy)
  dz := 7
  gl.Begin(gl.QUADS)
    gl.TexCoord2f(corner, 0)
    gl.Vertex3i(rv.dx, rv.dy, 0)
    gl.TexCoord2f(corner, -1)
    gl.Vertex3i(rv.dx, rv.dy, -dz)
    gl.TexCoord2f(0, -1)
    gl.Vertex3i(0, rv.dy, -dz)
    gl.TexCoord2f(0, 0)
    gl.Vertex3i(0, rv.dy, 0)

    gl.TexCoord2f(1, 0)
    gl.Vertex3i(rv.dx, 0, 0)
    gl.TexCoord2f(1, -1)
    gl.Vertex3i(rv.dx, 0, -dz)
    gl.TexCoord2f(corner, -1)
    gl.Vertex3i(rv.dx, rv.dy, -dz)
    gl.TexCoord2f(corner, 0)
    gl.Vertex3i(rv.dx, rv.dy, 0)
  gl.End()



  gl.Disable(gl.TEXTURE_2D)
  gl.Color4f(1, 0, 1, 0.9)
  gl.LineWidth(3.0)
  gl.Begin(gl.LINES)
  for i := float32(0); i < float32(rv.dx); i += 1.0 {
    gl.Vertex2f(i, 0)
    gl.Vertex2f(i, float32(rv.dy))
  }
  for j := float32(0); j < float32(rv.dy); j += 1.0 {
    gl.Vertex2f(0, j)
    gl.Vertex2f(float32(rv.dx), j)
  }
  gl.End()

  gl.Enable(gl.TEXTURE_2D)
  gl.Color4d(1, 1, 1, 1)
  // Draw a furniture tile
  rv.cube.texture.Bind(gl.TEXTURE_2D)
  gl.PushMatrix()
  gl.LoadIdentity()

  sort.Sort(rv.furn)
  for _,f := range rv.furn {
  near_x := f.Pos().X
  near_y := f.Pos().Y
  furn_dx := f.Dims().X
  furn_dy := f.Dims().Y
  leftx,_,_ := rv.boardToModelview(near_x, near_y + furn_dy)
  rightx,_,_ := rv.boardToModelview(near_x + furn_dx, near_y)
  _,boty,_ := rv.boardToModelview(near_x, near_y)
    f.Render(mathgl.Vec2{leftx, boty}, rightx - leftx)
  }
  gl.PopMatrix()

  for i := range rv.flattened_positions {
    v := rv.flattened_positions[i]
    rv.flattened_drawables[i].Render(v.X, v.Y, 0, 1.0)
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
