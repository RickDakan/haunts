package house

import (
  "haunts/base"
  "haunts/texture"
  "github.com/arbaal/mathgl"
  "gl"
  "path/filepath"
  "os"
  "sort"
)

var (
  furniture_registry map[string]*furnitureDef
)

func init() {
  furniture_registry = make(map[string]*furnitureDef)
}

func MakeFurniture(name string) *Furniture {
  f := Furniture{ Defname: name }
  f.Load()
  return &f
}

func GetAllFurnitureNames() []string {
  var names []string
  for name := range furniture_registry {
    names = append(names, name)
  }
  sort.Strings(names)
  return names
}

func LoadAllFurnitureInDir(dir string) {
  filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
    if !info.IsDir() {
      if len(info.Name()) >= 5 && info.Name()[len(info.Name()) - 5 : ] == ".json" {
        var f furnitureDef
        err := base.LoadJson(path, &f)
        if err == nil {
          f.abs_texture_path = filepath.Clean(filepath.Join(path, f.Texture_path))
          furniture_registry[f.Name] = &f
        }
      }
    }
    return nil
  })
}

type Furniture struct {
  Defname string
  *furnitureDef
  FurnitureInst
}

func (f *Furniture) Load() {
  f.furnitureDef = furniture_registry[f.Defname]
  if f.furnitureDef.texture_data == nil {
    f.furnitureDef.texture_data = texture.LoadFromPath(f.furnitureDef.abs_texture_path)
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

  // Path to the texture - relative to the location of this file
  Texture_path string

  // Absolute path to texture
  abs_texture_path string

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
