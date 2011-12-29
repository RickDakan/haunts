package house

import (
  "glop/gui"
  "glop/gin"
  "glop/sprite"
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

  // In case the size of the room changes we will need to update the matrices
  size RoomSize

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
    WallTexture *WallTexture
  }

  Selected struct {
    Cells map[CellPos]bool
  }

  floor *texture.Data
  wall *texture.Data

  // This tells us what to highlight based on the mouse position
  edit_mode editMode
}

func (rv *RoomViewer) SetEditMode(mode editMode) {
  rv.edit_mode = mode
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
  rv.size = rv.room.Size
  rv.makeMat()
  rv.Request_dims.Dx = 100
  rv.Request_dims.Dy = 100
  rv.Ex = true
  rv.Ey = true

  rv.Selected.Cells = make(map[CellPos]bool)

  return &rv
}

func (rv *RoomViewer) AdjAngle(ang float32) {
  rv.angle = ang
  rv.makeMat()
}

func (rv *RoomViewer) Drag(dx,dy float64) {
  x,y,_ := rv.boardToModelview(rv.fx, rv.fy)
  x += float32(dx)
  y += float32(dy)
  rv.fx, rv.fy, _ = rv.modelviewToBoard(x, y)
  rv.fx = clamp(rv.fx, -2, float32(rv.room.Size.Dx) + 2)
  rv.fy = clamp(rv.fy, -2, float32(rv.room.Size.Dy) + 2)
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
  m.RotationY(-math.Pi/2)
  rv.right_wall_mat.Multiply(&m)
  m.Scaling(1, 1, 1)
  rv.right_wall_mat.Multiply(&m)
  m.Translation(0, 0, -float32(rv.room.Size.Dx))
  rv.right_wall_mat.Multiply(&m)
  swap_x_y := mathgl.Mat4 {
    0,1,0,0,
    1,0,0,0,
    0,0,1,0,
    0,0,0,1,
  }
  rv.right_wall_mat.Multiply(&swap_x_y)
  rv.right_wall_imat.Assign(&rv.right_wall_mat)
  rv.right_wall_imat.Inverse()
}

// Transforms a cursor position in window coordinates to board coordinates.
func (rv *RoomViewer) WindowToBoard(wx, wy int) (float32, float32) {
  fx,fy,fdist := rv.modelviewToBoard(float32(wx), float32(wy))
  lbx,lby,ldist := rv.modelviewToLeftWall(float32(wx), float32(wy))
  rbx,rby,rdist := rv.modelviewToRightWall(float32(wx), float32(wy))
  if fdist < ldist && fdist < rdist {
    if fx > float32(rv.room.Size.Dx) {
      fx = float32(rv.room.Size.Dx)
    }
    if fy > float32(rv.room.Size.Dy) {
      fy = float32(rv.room.Size.Dy)
    }
    return fx, fy
  }
  if ldist < rdist {
    return lbx, lby
  }
  return rbx, rby
}

func (rv *RoomViewer) BoardToWindow(bx,by float32) (float32, float32) {
  fx,fy,fz := rv.boardToModelview(float32(bx), float32(by))
  lbx,lby,lz := rv.leftWallToModelview(float32(bx), float32(by))
  rbx,rby,rz := rv.rightWallToModelview(float32(bx), float32(by))
  if fz < lz && fz < rz {
    return fx, fy
  }
  if lz < rz {
    return lbx, lby
  }
  return rbx, rby
}

func (rv *RoomViewer) modelviewToLeftWall(mx, my float32) (x,y,dist float32) {
  mz := d2p(rv.left_wall_mat, mathgl.Vec3{mx, my, 0}, mathgl.Vec3{0,0,1})
  v := mathgl.Vec4{X: mx, Y: my, Z: mz, W: 1}
  v.Transform(&rv.left_wall_imat)
  if v.X > float32(rv.room.Size.Dx) {
    v.X = float32(rv.room.Size.Dx)
  }
  return v.X, v.Y + float32(rv.room.Size.Dy), mz
}

func (rv *RoomViewer) modelviewToRightWall(mx, my float32) (x,y,dist float32) {
  mz := d2p(rv.right_wall_mat, mathgl.Vec3{mx, my, 0}, mathgl.Vec3{0,0,1})
  v := mathgl.Vec4{X: mx, Y: my, Z: mz, W: 1}
  v.Transform(&rv.right_wall_imat)
  if v.Y > float32(rv.room.Size.Dy) {
    v.Y = float32(rv.room.Size.Dy)
  }
  return v.X + float32(rv.room.Size.Dx), v.Y, mz
}

func (rv *RoomViewer) leftWallToModelview(bx,by float32) (x, y, z float32) {
  v := mathgl.Vec4{X: bx, Y: by - float32(rv.room.Size.Dy), W: 1}
  v.Transform(&rv.left_wall_mat)
  return v.X, v.Y, v.Z
}

func (rv *RoomViewer) rightWallToModelview(bx,by float32) (x, y, z float32) {
  v := mathgl.Vec4{X: bx - float32(rv.room.Size.Dx), Y: by, W: 1}
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

func (rv *RoomViewer) modelviewToBoard(mx, my float32) (x,y,dist float32) {
  mz := d2p(rv.mat, mathgl.Vec3{mx, my, 0}, mathgl.Vec3{0,0,1})
//  mz := (my - float32(rv.Render_region.Y+rv.Render_region.Dy/2)) * float32(math.Tan(float64(rv.angle*math.Pi/180)))
  v := mathgl.Vec4{X: mx, Y: my, Z: mz, W: 1}
  v.Transform(&rv.imat)
  return v.X, v.Y, mz
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
  rv.fx = clamp(rv.fx, -2, float32(rv.room.Size.Dx) + 2)
  rv.fy = clamp(rv.fy, -2, float32(rv.room.Size.Dy) + 2)
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

func (rv *RoomViewer) drawWall() {
  gl.Enable(gl.STENCIL_TEST)
  defer gl.Disable(gl.STENCIL_TEST)

  // Right wall
  gl.ClearStencil(0)
  gl.Clear(gl.STENCIL_BUFFER_BIT)
  gl.StencilFunc(gl.ALWAYS, 1, 1)
  gl.StencilOp(gl.REPLACE, gl.REPLACE, gl.REPLACE)
  gl.Disable(gl.TEXTURE_2D)
  dz := 7
  gl.Begin(gl.QUADS)
    gl.Vertex3i(rv.room.Size.Dx, 0, 0)
    gl.Vertex3i(rv.room.Size.Dx, 0, -dz)
    gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dy, -dz)
    gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dy, 0)
  gl.End()
  gl.StencilFunc(gl.EQUAL, 1, 1)
  gl.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)

  gl.Enable(gl.TEXTURE_2D)
  rv.wall.Bind()
  gl.Color4f(1, 1, 1, 1)
  corner := float32(rv.room.Size.Dx) / float32(rv.room.Size.Dx + rv.room.Size.Dy)
  gl.Begin(gl.QUADS)
    gl.TexCoord2f(1, 0)
    gl.Vertex3i(rv.room.Size.Dx, 0, 0)
    gl.TexCoord2f(1, -1)
    gl.Vertex3i(rv.room.Size.Dx, 0, -dz)
    gl.TexCoord2f(corner, -1)
    gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dy, -dz)
    gl.TexCoord2f(corner, 0)
    gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dy, 0)
  gl.End()

  gl.PushMatrix()
  gl.LoadIdentity()
  gl.MultMatrixf(&rv.right_wall_mat[0])
  var texs []WallTexture
  if rv.Temp.WallTexture != nil {
    texs = append(texs, *rv.Temp.WallTexture)
  }
  for _,tex := range rv.room.WallTextures {
    texs = append(texs, *tex)
  }
  for i,tex := range texs {
    dx, dy := float32(rv.room.Size.Dx), float32(rv.room.Size.Dy)
    if tex.Y > dy {
      tex.X, tex.Y = dx + tex.Y - dy, dy + dx - tex.X
    }
    if tex.X > dx {
      tex.Rot -= 3.1415926535 / 2
    }
    tex.X -= dx
    if rv.Temp.WallTexture != nil && i == 0 {
      gl.Color4f(1, 0.7, 0.7, 0.7)
    } else {
      gl.Color4f(1, 1, 1, 1)
    }
    tex.Render()
  }
  gl.PopMatrix()

  // Left wall
  gl.ClearStencil(0)
  gl.Clear(gl.STENCIL_BUFFER_BIT)
  gl.StencilFunc(gl.ALWAYS, 1, 1)
  gl.StencilOp(gl.REPLACE, gl.REPLACE, gl.REPLACE)
  gl.Disable(gl.TEXTURE_2D)
  gl.Begin(gl.QUADS)
    gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dy, 0)
    gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dy, -dz)
    gl.Vertex3i(0, rv.room.Size.Dy, -dz)
    gl.Vertex3i(0, rv.room.Size.Dy, 0)
  gl.End()
  gl.StencilFunc(gl.EQUAL, 1, 1)
  gl.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)

  rv.wall.Bind()
  gl.Enable(gl.TEXTURE_2D)
  gl.Color4f(1, 1, 1, 1)
  gl.Begin(gl.QUADS)
    gl.TexCoord2f(corner, 0)
    gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dy, 0)
    gl.TexCoord2f(corner, -1)
    gl.Vertex3i(rv.room.Size.Dx, rv.room.Size.Dy, -dz)
    gl.TexCoord2f(0, -1)
    gl.Vertex3i(0, rv.room.Size.Dy, -dz)
    gl.TexCoord2f(0, 0)
    gl.Vertex3i(0, rv.room.Size.Dy, 0)
  gl.End()

  gl.PushMatrix()
  gl.LoadIdentity()
  gl.MultMatrixf(&rv.left_wall_mat[0])
  for i,tex := range texs {
    dx, dy := float32(rv.room.Size.Dx), float32(rv.room.Size.Dy)
    if tex.X > dx {
      tex.X, tex.Y = dx + dy - tex.Y, dy + tex.X - dx
    }
    tex.Y -= dy
    if rv.Temp.WallTexture != nil && i == 0 {
      gl.Color4f(1, 0.7, 0.7, 0.7)
    } else {
      gl.Color4f(1, 1, 1, 1)
    }
    tex.Render()
  }
  gl.PopMatrix()

}

func (rv *RoomViewer) drawFloor() {
  gl.Enable(gl.STENCIL_TEST)
  defer gl.Disable(gl.STENCIL_TEST)

  gl.ClearStencil(0)
  gl.Clear(gl.STENCIL_BUFFER_BIT)
  gl.StencilFunc(gl.ALWAYS, 1, 1)
  gl.StencilOp(gl.REPLACE, gl.REPLACE, gl.REPLACE)
  gl.Disable(gl.TEXTURE_2D)
  gl.Begin(gl.QUADS)
    gl.Vertex2i(0, 0)
    gl.Vertex2i(0, rv.room.Size.Dy)
    gl.Vertex2i(rv.room.Size.Dx, rv.room.Size.Dy)
    gl.Vertex2i(rv.room.Size.Dx, 0)
  gl.End()
  gl.StencilFunc(gl.EQUAL, 1, 1)
  gl.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)

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
  {
//    clip := gui.Region{ gui.Point{0, 0}, gui.Dims{rv.room.Size.Dx, rv.room.Size.Dy} }
//    clip.PushClipPlanes()
    var texs []WallTexture
    if rv.Temp.WallTexture != nil {
      texs = append(texs, *rv.Temp.WallTexture)
    }
    for _,tex := range rv.room.WallTextures {
      texs = append(texs, *tex)
    }
    for i,tex := range texs {
      if tex.X >= float32(rv.room.Size.Dx) {
        tex.Rot -= 3.1415926535 / 2
      }
      if rv.Temp.WallTexture != nil && i == 0 {
        gl.Color4f(1, 0.7, 0.7, 0.7)
      } else {
        gl.Color4f(1, 1, 1, 1)
      }
      tex.Render()
    }
//    clip.PopClipPlanes()
  }

  if rv.edit_mode == editCells {
    gl.Disable(gl.TEXTURE_2D)
    gl.Color4d(0.3, 1, 0.3, 0.7)
    gl.Begin(gl.QUADS)
      for pos := range rv.Selected.Cells {
        gl.Vertex2i(pos.X, pos.Y)
        gl.Vertex2i(pos.X, pos.Y + 1)
        gl.Vertex2i(pos.X + 1, pos.Y + 1)
        gl.Vertex2i(pos.X + 1, pos.Y)
      }
    gl.End()
  }

  gl.Disable(gl.TEXTURE_2D)
  gl.Color4f(1, 0, 1, 0.9)
  if rv.edit_mode == editCells {
    gl.LineWidth(0.02 * rv.zoom)
  } else {
    gl.LineWidth(0.05 * rv.zoom)
  }
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

  if rv.edit_mode == editCells {
    gl.Disable(gl.TEXTURE_2D)
    gl.Color4d(1, 0, 0, 1)
    gl.LineWidth(0.05 * rv.zoom)
    gl.Begin(gl.LINES)
      for _,f := range rv.room.Furniture {
        x,y := f.Pos()
        dx,dy := f.Dims()
        gl.Vertex2i(x, y)
        gl.Vertex2i(x, y + dy)

        gl.Vertex2i(x, y + dy)
        gl.Vertex2i(x + dx, y + dy)

        gl.Vertex2i(x + dx, y + dy)
        gl.Vertex2i(x + dx, y)

        gl.Vertex2i(x + dx, y)
        gl.Vertex2i(x, y)
      }
    gl.End()
  }

  gl.Disable(gl.STENCIL_TEST)
  if rv.edit_mode == editCells {
    for i := range rv.room.Cell_data {
      for j := range rv.room.Cell_data[i] {
        rv.room.Cell_data[i][j].Render(i, j, rv.room.Size.Dx, rv.room.Size.Dy)
      }
    }
  }

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

  rv.drawWall()
  rv.drawFloor()

  gl.Enable(gl.TEXTURE_2D)
  gl.Color4d(1, 1, 1, 1)
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
  if rv.size != rv.room.Size {
    rv.size = rv.room.Size
    rv.makeMat()
  }
  mx,my := rv.WindowToBoard(gin.In().GetCursor("Mouse").Point())
  rv.mx = int(mx)
  rv.my = int(my)
}
