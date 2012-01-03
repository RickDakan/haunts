package house

import (
  "haunts/base"
  "haunts/texture"
  "github.com/arbaal/mathgl"
  "gl"
)

func init() {
  base.RegisterRegistry("furniture", make(map[string]*furnitureDef))
}

func MakeFurniture(name string) *Furniture {
  f := Furniture{ Defname: name }
  f.Load()
  return &f
}

func GetAllFurnitureNames() []string {
  return base.GetAllNamesInRegistry("furniture")
}

func LoadAllFurnitureInDir(dir string) {
  base.RegisterAllObjectsInDir("furniture", dir, ".json", "json")
}

type Furniture struct {
  Defname string
  *furnitureDef
  FurnitureInst
}

func (f *Furniture) Load() {
  base.LoadObject("furniture", f)
  if f.furnitureDef.texture_data == nil {
    f.furnitureDef.texture_data = texture.LoadFromPath(f.furnitureDef.Texture_path)
  }
}

// Changes the position of this object such that it fits within the specified
// dimensions, if possible
func (f *Furniture) Constrain(dx,dy int) {
  if f.X + f.Dx > dx {
    f.X += dx - f.X + f.Dx
  }
  if f.Y + f.Dy > dy {
    f.Y += dy - f.Y + f.Dy
  }
}

// This data is what differentiates different instances of the same piece of
// furniture
type FurnitureInst struct {
  // Position of this object in board coordinates.
  X,Y int
}

func (f *FurnitureInst) Pos() (int, int) {
  return f.X, f.Y
}

// All instances of the same piece of furniture have this data in common
type furnitureDef struct {
  // Name of the object - should be unique among all furniture
  Name string

  // Dimensions of the object in cells
  Dx,Dy int

  // Path to the texture - on disk this is stored as a path that is relative
  // to the location of the file that this def is stored in.  When it is loaded
  // it is converted to an absolute path
  Texture_path string `registry:"path"`

  // The texture itself
  texture_data *texture.Data
}

func (f *furnitureDef) Dims() (int, int) {
  return f.Dx, f.Dy
}

func (f *furnitureDef) RenderDims(pos mathgl.Vec2, width float32) {
  dy := width * float32(f.texture_data.Dy) / float32(f.texture_data.Dx)

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

func (f *furnitureDef) Render(pos mathgl.Vec2, width float32) {
  dy := width * float32(f.texture_data.Dy) / float32(f.texture_data.Dx)
  gl.Enable(gl.TEXTURE_2D)
  f.texture_data.Bind()
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
