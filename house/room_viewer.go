package house

import (
  "glop/gui"
  "glop/gin"
  "glop/sprite"
  "glop/util/algorithm"
  "gl"
  "math"
  "github.com/arbaal/mathgl"
  "haunts/texture"
)

type RectObject interface {
  // Position in board coordinates
  Pos() (int, int)

  // Dimensions in board coordinates
  Dims() (int, int)

  Render(pos mathgl.Vec2, width float32)
  RenderDims(pos mathgl.Vec2, width float32)
}


type rectObjectArray []RectObject
func (r rectObjectArray) Order() rectObjectArray {
  var nr rectObjectArray
  if len(r) == 0 {
    return nil
  }
  if len(r) == 1 {
    nr = append(nr, r[0])
    return nr
  }

  minx,miny := r[0].Pos()
  maxx,maxy := r[0].Pos()
  for i := range r {
    x,y := r[i].Pos()
    if x < minx { minx = x }
    if y < miny { miny = y }
    if x > maxx { maxx = x }
    if y > maxy { maxy = y }
  }

  // check for an x-divide
  var low,high rectObjectArray
  for divx := minx; divx <= maxx; divx++ {
    low = low[0:0]
    high = high[0:0]
    for i := range r {
      x,_ := r[i].Pos()
      dx,_ := r[i].Dims()
      if x >= divx {
        high = append(high, r[i])
      }
      if x + dx - 1 < divx {
        low = append(low, r[i])
      }
    }
    if len(low) + len(high) == len(r) && len(low) >= 1 && len(high) >= 1 {
      low = low.Order()
      for i := range low {
        nr = append(nr, low[i])
      }
      high = high.Order()
      for i := range high {
        nr = append(nr, high[i])
      }
      return nr
    }
  }

  // check for a y-divide
  for divy := miny; divy <= maxy; divy++ {
    low = low[0:0]
    high = high[0:0]
    for i := range r {
      _,y := r[i].Pos()
      _,dy := r[i].Dims()
      if y >= divy {
        high = append(high, r[i])
      }
      if y + dy - 1 < divy {
        low = append(low, r[i])
      }
    }
    if len(low) + len(high) == len(r) && len(low) >= 1 && len(high) >= 1 {
      low = low.Order()
      for i := range low {
        nr = append(nr, low[i])
      }
      high = high.Order()
      for i := range high {
        nr = append(nr, high[i])
      }
      return nr
    }
  }
  for i := range r {
    nr = append(nr, r[i])
  }
  return nr
}
func (r rectObjectArray) Less(i,j int) bool {
  ix,iy := r[i].Pos()
  jdx,jdy := r[j].Dims()
  jx,jy := r[j].Pos()
  jx2 := jx + jdx - 1
  jy2 := jy + jdy - 1
  return jx2 < ix || (!(jx2 < ix) && jy2 < iy)
}
func (r rectObjectArray) LessX(i,j int) bool {
  ix,_ := r[i].Pos()
  jdx,_ := r[j].Dims()
  jx,_ := r[j].Pos()
  jx2 := jx + jdx - 1
  return jx2 < ix
}
func (r rectObjectArray) LessY(i,j int) bool {
  _,iy := r[i].Pos()
  _,jdy := r[j].Dims()
  _,jy := r[j].Pos()
  jy2 := jy + jdy - 1
  return jy2 < iy
}

type editMode int
const (
  editNothing editMode = iota
  editFurniture
  editWallTextures
  editCells
)

type RoomViewer struct {
  gui.Childless
  gui.EmbeddedWidget
  gui.BasicZone
  gui.NonFocuser
  gui.NonResponder
  gui.NonThinker

  room *Room

  // All events received by the viewer are passed to the handler
  handler gin.EventHandler

  // Focus, in map coordinates
  fx, fy float32

  // Mouse position, in board coordinates
  mx, my int

  // The viewing angle, 0 means the map is viewed head-on, 90 means the map is viewed
  // on its edge (i.e. it would not be visible)
  angle float32

  // Zoom factor, 1.0 is standard
  zoom float32

  // The modelview matrix that is sent to opengl.  Updated any time focus, zoom, or viewing
  // angle changes
  mat mathgl.Mat4
  left_wall_mat mathgl.Mat4
  right_wall_mat mathgl.Mat4

  // Inverse of mat
  imat mathgl.Mat4
  left_wall_imat mathgl.Mat4
  right_wall_imat mathgl.Mat4

  // All drawables that will be drawn parallel to the window
  upright_drawables []sprite.ZDrawable
  upright_positions []mathgl.Vec3

  // All drawables that will be drawn on the surface of the viewer
  flattened_drawables []sprite.ZDrawable
  flattened_positions []mathgl.Vec3

  Temp struct {
    Furniture *Furniture
  }

  temp_wall_texture *WallTexture
  wall_textures []*WallTexture

  floor *texture.Data
  wall *texture.Data

  // This tells us what to highlight based on the mouse position
  edit_mode editMode

  selected_walls map[int]bool
}

func (rv *RoomViewer) SetEditMode(mode editMode) {
  rv.edit_mode = mode
}

func (rv *RoomViewer) SetTempWallTexture(wt *WallTexture) {
  rv.RemoveWallTexture(rv.temp_wall_texture)
  rv.temp_wall_texture = wt
  if rv.temp_wall_texture != nil {
    rv.AddWallTexture(rv.temp_wall_texture)
  }
}

func (rv *RoomViewer) AddWallTexture(wt *WallTexture) {
  rv.wall_textures = append(rv.wall_textures, wt)
}

func (rv *RoomViewer) RemoveWallTexture(wt *WallTexture) {
  rv.wall_textures = algorithm.Choose(rv.wall_textures, func(a interface{}) bool { return a.(*WallTexture) != wt }).([]*WallTexture)
}

func (rv *RoomViewer) SelectWallTextureAt(wx,wy int) *WallTexture {
  pos,height := rv.WindowToWall(wx, wy)
  rv.temp_wall_texture = rv.wallTextureAt(pos, height)
  return rv.temp_wall_texture
}

func (rv *RoomViewer) wallTextureAt(pos,height float32) *WallTexture {
  sum := float32(rv.room.Size.Dx + rv.room.Size.Dy)
  pos *= sum
  for _,tex := range rv.wall_textures {
    dx := float32(tex.texture_data.Dx) / 100
    dy := float32(tex.texture_data.Dy) / 100
    tpos := tex.Pos * sum
    if pos >= tpos - dx && pos < tpos + dx && height >= tex.Height - dy && height < tex.Height + dy {
      return tex
    }
  }
  return nil
}

func (rv *RoomViewer) ReloadFloor(path string) {
  rv.floor = texture.LoadFromPath(path)
}

func (rv *RoomViewer) ReloadWall(path string) {
  rv.wall = texture.LoadFromPath(path)
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

func MakeRoomViewer(room *Room, angle float32) *RoomViewer {
  var rv RoomViewer
  rv.EmbeddedWidget = &gui.BasicWidget{CoreWidget: &rv}
  rv.room = room
  rv.angle = angle
  rv.fx = float32(rv.room.Size.Dx / 2)
  rv.fy = float32(rv.room.Size.Dy / 2)
  rv.Zoom(1)
  rv.makeMat()
  rv.Request_dims.Dx = 100
  rv.Request_dims.Dy = 100
  rv.Ex = true
  rv.Ey = true
  rv.selected_walls = make(map[int]bool)

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

  // Also make the mats for the left and right walls based on this mat
  rv.left_wall_mat.Assign(&rv.mat)
  m.RotationX(-math.Pi/2)
  rv.left_wall_mat.Multiply(&m)
  m.Translation(0, 0, float32(rv.room.Size.Dy))
  rv.left_wall_mat.Multiply(&m)
  rv.left_wall_imat.Assign(&rv.left_wall_mat)
  rv.left_wall_imat.Inverse()

  rv.right_wall_mat.Assign(&rv.mat)
  m.RotationX(-math.Pi/2)
  rv.right_wall_mat.Multiply(&m)
  m.RotationY(math.Pi/2)
  rv.right_wall_mat.Multiply(&m)
  m.Scaling(-1, 1, 1)
  rv.right_wall_mat.Multiply(&m)
  m.Translation(0, 0, float32(rv.room.Size.Dx))
  rv.right_wall_mat.Multiply(&m)
  rv.right_wall_imat.Assign(&rv.right_wall_mat)
  rv.right_wall_imat.Inverse()
}

// Transforms a cursor position in window coordinates to board coordinates.  Does not check
// to make sure that the values returned represent a valid position on the board.
func (rv *RoomViewer) WindowToBoard(wx, wy int) (float32, float32) {
  mx := float32(wx)
  my := float32(wy)
  return rv.modelviewToBoard(mx, my)
}

func (rv *RoomViewer) WindowToWall(wx, wy int) (pos, height float32) {
  lbx,lby,ldist := rv.modelviewToLeftWall(float32(wx), float32(wy))
  rbx,rby,rdist := rv.modelviewToRightWall(float32(wx), float32(wy))
  sum := float32(rv.room.Size.Dx + rv.room.Size.Dy)
  if ldist < rdist {
    pos = lbx / sum
    height = lby
  } else {
    pos = (sum - rbx) / sum
    height = rby
  }
  return
}

func (rv *RoomViewer) modelviewToLeftWall(mx, my float32) (x,y,dist float32) {
  mz := d2p(rv.left_wall_mat, mathgl.Vec3{mx, my, 0}, mathgl.Vec3{0,0,1})
  v := mathgl.Vec4{X: mx, Y: my, Z: mz, W: 1}
  v.Transform(&rv.left_wall_imat)
  return v.X, v.Y, mz
}

func (rv *RoomViewer) modelviewToRightWall(mx, my float32) (x,y,dist float32) {
  mz := d2p(rv.right_wall_mat, mathgl.Vec3{mx, my, 0}, mathgl.Vec3{0,0,1})
  v := mathgl.Vec4{X: mx, Y: my, Z: mz, W: 1}
  v.Transform(&rv.right_wall_imat)
  return v.X, v.Y, mz
}

func (rv *RoomViewer) leftWallToModelview(bx,by float32) (x, y, z float32) {
  v := mathgl.Vec4{X: bx, Y: by, W: 1}
  v.Transform(&rv.left_wall_mat)
  return v.X, v.Y, v.Z
}

func (rv *RoomViewer) rightWallToModelview(bx,by float32) (x, y, z float32) {
  v := mathgl.Vec4{X: bx, Y: by, W: 1}
  v.Transform(&rv.right_wall_mat)
  return v.X, v.Y, v.Z
}

func d2p(mat mathgl.Mat4, point,ray mathgl.Vec3) float32{
  var sub mathgl.Vec3
  sub.X = mat[12]
  sub.Y = mat[13]
  sub.Z = mat[14]
  mat[12],mat[13],mat[14] = 0,0,0
  point.Subtract(&sub)
  point.Scale(-1)
  ray.Normalize()
  dist := point.Dot(mat.GetForwardVec3())

  var n,r mathgl.Vec3
  n.X = mat.GetForwardVec3().X
  n.Y = mat.GetForwardVec3().Y
  n.Z = mat.GetForwardVec3().Z
  r.Assign(&ray)
  cos := n.Dot(&r) / (n.Length() * r.Length())
  R := dist / float32(math.Sin(math.Pi/2 - math.Acos(float64(cos))))
  return R
}

func (rv *RoomViewer) modelviewToBoard(mx, my float32) (float32, float32) {
  mz := d2p(rv.mat, mathgl.Vec3{mx, my, 0}, mathgl.Vec3{0,0,1})
//  mz := (my - float32(rv.Render_region.Y+rv.Render_region.Dy/2)) * float32(math.Tan(float64(rv.angle*math.Pi/180)))
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
  rv.fx = clamp(rv.fx, 0, float32(rv.room.Size.Dx))
  rv.fy = clamp(rv.fy, 0, float32(rv.room.Size.Dy))
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
  gl.Vertex2i(-1, rv.room.Size.Dy+1)
  gl.Vertex2i(rv.room.Size.Dx+1, rv.room.Size.Dy+1)
  gl.Vertex2i(rv.room.Size.Dx+1, -1)
  gl.End()


  // Draw the wall
  rv.wall.Bind()
  gl.Enable(gl.TEXTURE_2D)
  gl.Color4f(1, 1, 1, 1)
  corner := float32(rv.room.Size.Dx) / float32(rv.room.Size.Dx + rv.room.Size.Dy)
  dz := 7
  gl.Begin(gl.QUADS)
    gl.TexCoord2f(corner, 0)
    gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dy, 0)
    gl.TexCoord2f(corner, -1)
    gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dy, -dz)
    gl.TexCoord2f(0, -1)
    gl.Vertex3i(0, rv.room.Size.Dy, -dz)
    gl.TexCoord2f(0, 0)
    gl.Vertex3i(0, rv.room.Size.Dy, 0)

    gl.TexCoord2f(1, 0)
    gl.Vertex3i(rv.room.Size.Dx, 0, 0)
    gl.TexCoord2f(1, -1)
    gl.Vertex3i(rv.room.Size.Dx, 0, -dz)
    gl.TexCoord2f(corner, -1)
    gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dy, -dz)
    gl.TexCoord2f(corner, 0)
    gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dy, 0)
  gl.End()
  for _,tex := range rv.wall_textures {
    tex.Render(rv.room.Size.Dx, rv.room.Size.Dy)
  }
  gl.Disable(gl.TEXTURE_2D)
  for v := range rv.selected_walls {
    if v < rv.room.Size.Dx {
      gl.Begin(gl.QUADS)
        gl.Color4f(1, 0, 0, 0.0)
        gl.Vertex3i(v+1, rv.room.Size.Dy, 0)
        gl.Vertex3i(v, rv.room.Size.Dy, 0)
        gl.Color4f(1, 0, 0, 0.5)
        gl.Vertex3i(v, rv.room.Size.Dy, -10)
        gl.Vertex3i(v+1, rv.room.Size.Dy, -10)
      gl.End()
    } else {
      gl.Begin(gl.QUADS)
        gl.Color4f(1, 0, 0, 0.0)
        gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dx + rv.room.Size.Dy - v - 1, 0)
        gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dx + rv.room.Size.Dy - v, 0)
        gl.Color4f(1, 0, 0, 0.5)
        gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dx + rv.room.Size.Dy - v, -10)
        gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dx + rv.room.Size.Dy - v - 1, -10)
      gl.End()
    }
  }

  // Draw the floor
  gl.Enable(gl.TEXTURE_2D)
  rv.floor.Bind()
  gl.Color4d(1.0, 1.0, 1.0, 1.0)
  gl.Begin(gl.QUADS)
    gl.TexCoord2i(0, 0)
    gl.Vertex2i(0, 0)
    gl.TexCoord2i(0, -1)
    gl.Vertex2i(0, rv.room.Size.Dy)
    gl.TexCoord2i(1, -1)
    gl.Vertex2i(rv.room.Size.Dx, rv.room.Size.Dy)
    gl.TexCoord2i(1, 0)
    gl.Vertex2i(rv.room.Size.Dx, 0)
  gl.End()
  // If we're in cell-select mode we can highlight the cell our mouse is over
  // if rv.edit_mode == selectCells && rv.mx >= 0 && rv.mx < rv.room.Size.Dx && rv.my >= 0 && rv.my < rv.room.Size.Dy {
  //   gl.Disable(gl.TEXTURE_2D)
  //   gl.Color4d(0.5, 1, 0.5, 1)
  //   gl.Begin(gl.QUADS)
  //     gl.Vertex2i(rv.mx, rv.my)
  //     gl.Vertex2i(rv.mx, rv.my + 1)
  //     gl.Vertex2i(rv.mx + 1, rv.my + 1)
  //     gl.Vertex2i(rv.mx + 1, rv.my)
  //   gl.End()
  // }

  gl.Disable(gl.TEXTURE_2D)
  gl.Color4f(1, 0, 1, 0.9)
  gl.LineWidth(3.0)
  gl.Begin(gl.LINES)
  for i := float32(0); i < float32(rv.room.Size.Dx); i += 1.0 {
    gl.Vertex2f(i, 0)
    gl.Vertex2f(i, float32(rv.room.Size.Dy))
  }
  for j := float32(0); j < float32(rv.room.Size.Dy); j += 1.0 {
    gl.Vertex2f(0, j)
    gl.Vertex2f(float32(rv.room.Size.Dx), j)
  }
  gl.End()

  gl.Enable(gl.TEXTURE_2D)
  gl.Color4d(1, 1, 1, 1)
  // Draw a furniture tile
  gl.PushMatrix()
  gl.LoadIdentity()

  var furn rectObjectArray
  for _,f := range rv.room.Furniture {
    furn = append(furn, f)
  }
  if rv.Temp.Furniture != nil {
    furn = append(furn, rv.Temp.Furniture)
  }
  furn = furn.Order()
  var furn_over *Furniture
  for i := len(furn) - 1; i >= 0; i-- {
    f := furn[i]
    near_x,near_y := f.Pos()
    furn_dx,furn_dy := f.Dims()
    leftx,_,_ := rv.boardToModelview(float32(near_x), float32(near_y + furn_dy))
    rightx,_,_ := rv.boardToModelview(float32(near_x + furn_dx), float32(near_y))
    _,boty,_ := rv.boardToModelview(float32(near_x), float32(near_y))
    if f == rv.Temp.Furniture {
      gl.Color4d(1, 1, 1, 0.5)
    } else {
      if rv.edit_mode == editFurniture {
        if f == furn_over {
          gl.Color4d(0.5, 1, 0.5, 1)
        } else {
          gl.Color4d(1, 1, 1, 1)
        }
      } else {
        gl.Color4d(1, 1, 1, 0.4)
      }
    }
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

func (rv *RoomViewer) Think(*gui.Gui, int64) {
  mx,my := rv.WindowToBoard(gin.In().GetCursor("Mouse").Point())
  rv.mx = int(mx)
  rv.my = int(my)
}
