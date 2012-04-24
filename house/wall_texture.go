package house

import (
  "unsafe"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/mathgl"
  gl "github.com/chsc/gogl/gl21"
)

func MakeWallTexture(name string) *WallTexture {
  wt := WallTexture{ Defname: name }
  wt.Load()
  return &wt
}

func GetAllWallTextureNames() []string {
  return base.GetAllNamesInRegistry("wall_textures")
}

func LoadAllWallTexturesInDir(dir string) {
  base.RemoveRegistry("wall_textures")
  base.RegisterRegistry("wall_textures", make(map[string]*wallTextureDef))
  base.RegisterAllObjectsInDir("wall_textures", dir, ".json", "json")
}

func (wt *WallTexture) Load() {
  base.GetObject("wall_textures", wt)
}

type WallTexture struct {
  Defname string
  *wallTextureDef

  // Position of the texture in floor coordinates.  If these coordinates exceed
  // either the dx or dy of the room, then this texture will be drawn, at least
  // partially, on the wall.  The coordinates should not both exceed the
  // dimensions of the room.
  X,Y float32
  Rot float32

  // Whether or not to flip the texture about one of its axes
  Flip bool

  // If this is currently being dragged around it will be marked as temporary
  // so that it will be drawn differently
  temporary bool

  // we don't want to redo all of the vertex and index buffers unless we
  // need to, so we keep track of the position and size of the room when they
  // were made so we don't have to.
  gl struct {
    vbuffer uint32

    left_buffer uint32
    left_count gl.Sizei
    right_buffer uint32
    right_count gl.Sizei
    floor_buffer uint32
    floor_count gl.Sizei

    // for tracking whether the buffers are dirty
    x, y, rot float32
    flip bool
    room struct {
      x, y, dx, dy int
    }
  }
}

type wallTextureDef struct {
  // Name of this texture as it appears in the editor, should be unique among
  // all WallTextures
  Name string

  Texture texture.Object
}

func (wt *WallTexture) Color() (r,g,b,a byte) {
  if wt.temporary {
    return 127, 127, 255, 200
  }
  return 255, 255, 255, 255
}

func (wt *WallTexture) Render() {
  dx2 := float32(wt.Texture.Data().Dx()) / 100 / 2
  dy2 := float32(wt.Texture.Data().Dy()) / 100 / 2
  wt.Texture.Data().RenderAdvanced(float64(wt.X-dx2), float64(wt.Y-dy2), float64(2*dx2), float64(2*dy2), float64(wt.Rot), wt.Flip)
}



func (wt *WallTexture) setupGlStuff(room *Room) {
  dirty := false
  if wt.gl.room.x != room.X { dirty = true }
  if wt.gl.room.y != room.Y { dirty = true }
  if wt.gl.room.dx != room.Size.Dx { dirty = true }
  if wt.gl.room.dy != room.Size.Dy { dirty = true }
  if wt.gl.x != wt.X { dirty = true }
  if wt.gl.y != wt.Y { dirty = true }
  if wt.gl.rot != wt.Rot { dirty = true }
  if wt.gl.flip != wt.Flip { dirty = true }
  if wt.gl.vbuffer == 0 { dirty = true }
  if !dirty {
    return
  }
  base.Log().Printf("Wall texture setup gl")
  wt.gl.room.x = room.X
  wt.gl.room.y = room.Y
  wt.gl.room.dx = room.Size.Dx
  wt.gl.room.dy = room.Size.Dy
  wt.gl.x = wt.X
  wt.gl.y = wt.Y
  wt.gl.rot = wt.Rot
  wt.gl.flip = wt.Flip

  if wt.gl.vbuffer != 0 {
    gl.DeleteBuffers(1, &wt.gl.vbuffer)
    gl.DeleteBuffers(1, &wt.gl.left_buffer)
    gl.DeleteBuffers(1, &wt.gl.right_buffer)
    gl.DeleteBuffers(1, &wt.gl.floor_buffer)
    wt.gl.left_buffer = 0
    wt.gl.right_buffer = 0
    wt.gl.floor_buffer = 0
  }
  // dx := float32(room.Size.Dx)
  // dy := float32(room.Size.Dy)
  // var dz float32
  // if room.Wall.Data().Dx() > 0 {
  //   dz = -float32(room.Wall.Data().Dy() * (room.Size.Dx + room.Size.Dy)) / float32(room.Wall.Data().Dx())
  // }

  // Conveniently casted values
  frx := float32(room.X)
  fry := float32(room.Y)
  // frdx := float32(room.Size.Dx)
  // frdy := float32(room.Size.Dy)

  tdx := float32(wt.Texture.Data().Dx()) / 100
  tdy := float32(wt.Texture.Data().Dy()) / 100
  verts := []mathgl.Vec2{
    {-tdx / 2, -tdy / 2},
    {-tdx / 2,  tdy / 2},
    { tdx / 2,  tdy / 2},
    { tdx / 2, -tdy / 2},
  }
  var m mathgl.Mat3
  // m.RotationZ(wt.Rot)
  // for i := range verts {
  //   verts[i].Transform(&m)
  // }
  m.Translation(wt.X, wt.Y)
  for i := range verts {
    verts[i].Transform(&m)
  }
  vs := []roomVertex{
    {verts[0].X, verts[0].Y, 0, 1, -1, (verts[0].X + frx) / LosTextureSize, (verts[0].Y + fry) / LosTextureSize},
    {verts[1].X, verts[1].Y, 0, 0, 0, (verts[1].X + frx) / LosTextureSize, (verts[1].Y + fry) / LosTextureSize},
    {verts[2].X, verts[2].Y, 0, 1, 0, (verts[2].X + frx) / LosTextureSize, (verts[2].Y + fry) / LosTextureSize},
    {verts[3].X, verts[3].Y, 0, 1, 1, (verts[3].X + frx) / LosTextureSize, (verts[3].Y + fry) / LosTextureSize},
  }


  // vs := []roomVertex{
  //   // Walls
  //   // { 0,  dy,  0, 0, 1, lt_ury, lt_llx},
  //   // {dx,  dy,  0, c, 1, lt_ury, lt_urx},
  //   // {dx,   0,  0, 1, 1, lt_lly, lt_urx},
  //   // { 0,  dy, dz, 0, 0, lt_ury, lt_llx},
  //   // {dx,  dy, dz, c, 0, lt_ury, lt_urx},
  //   // {dx,   0, dz, 1, 0, lt_lly, lt_urx},

  //   // Floor
  //   { 0,  0, 0, 0, 1, lt_lly, lt_llx},
  //   { 0, dy, 0, 0, 0, lt_ury, lt_llx},
  //   {dx, dy, 0, 1, 0, lt_ury, lt_urx},
  //   {dx,  0, 0, 1, 1, lt_lly, lt_urx},
  // }
  gl.GenBuffers(1, &wt.gl.vbuffer)
  gl.BindBuffer(gl.ARRAY_BUFFER, wt.gl.vbuffer)
  size := int(unsafe.Sizeof(roomVertex{}))
  gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(size*len(vs)), gl.Pointer(&vs[0].x), gl.STATIC_DRAW)

  // left wall indices
  // is := []uint16{0, 3, 4, 0, 4, 1}
  // gl.GenBuffers(1, &room.left_buffer)
  // gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, room.left_buffer)
  // gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, gl.Sizeiptr(int(unsafe.Sizeof(is[0]))*len(is)), gl.Pointer(&is[0]), gl.STATIC_DRAW)

  // // right wall indices
  // is = []uint16{1, 4, 5, 1, 5, 2}
  // gl.GenBuffers(1, &room.right_buffer)
  // gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, room.right_buffer)
  // gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, gl.Sizeiptr(int(unsafe.Sizeof(is[0]))*len(is)), gl.Pointer(&is[0]), gl.STATIC_DRAW)

  // floor indices
  is := []uint16{0, 1, 3, 1, 2, 3}
  gl.GenBuffers(1, &wt.gl.floor_buffer)
  gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, wt.gl.floor_buffer)
  gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, gl.Sizeiptr(int(unsafe.Sizeof(is[0]))*len(is)), gl.Pointer(&is[0]), gl.STATIC_DRAW)
  wt.gl.floor_count = 6
}



