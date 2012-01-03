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
}

// Changes the position of this object such that it fits within the specified
// dimensions, if possible
func (f *Furniture) Constrain(dx,dy int) {
  cdx,cdy := f.Dims()
  if f.X + cdx > dx {
    f.X += dx - f.X + cdx
  }
  if f.Y + cdy > dy {
    f.Y += dy - f.Y + cdy
  }
}

// This data is what differentiates different instances of the same piece of
// furniture
type FurnitureInst struct {
  // Position of this object in board coordinates.
  X,Y int

  // Index into furnitureDef.Texture_paths
  Rotation int
}

func (f *FurnitureInst) Pos() (int, int) {
  return f.X, f.Y
}

func (f *Furniture) RotateLeft() {
  f.Rotation = (f.Rotation + 1) % len(f.Orientations)
}

func (f *Furniture) RotateRight() {
  f.Rotation = (f.Rotation - 1 + len(f.Orientations)) % len(f.Orientations)
}

type furnitureOrientation struct {
  Dx,Dy int
  Texture_path string `registry:"path"`
  texture_data *texture.Data
}

func (fo *furnitureOrientation) Load() {
  fo.texture_data = texture.LoadFromPath(fo.Texture_path)
}

// All instances of the same piece of furniture have this data in common
type furnitureDef struct {
  // Name of the object - should be unique among all furniture
  Name string

  // All available orientations for this piece of furniture
  Orientations []furnitureOrientation `registry:"autoload"`
}

func (f *Furniture) Dims() (int, int) {
  orientation := f.Orientations[f.Rotation]
  return orientation.Dx, orientation.Dy
}

func (f *Furniture) RenderDims(pos mathgl.Vec2, width float32) {
  orientation := f.Orientations[f.Rotation]
  dy := width * float32(orientation.texture_data.Dy) / float32(orientation.texture_data.Dx)

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

func (f *Furniture) Render(pos mathgl.Vec2, width float32) {
  orientation := f.Orientations[f.Rotation]
  dy := width * float32(orientation.texture_data.Dy) / float32(orientation.texture_data.Dx)
  gl.Enable(gl.TEXTURE_2D)
  orientation.texture_data.Bind()
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
