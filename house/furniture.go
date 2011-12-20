package house

import (
  "haunts/base"
  "haunts/texture"
  "github.com/arbaal/mathgl"
  "gl"
  "path/filepath"
  "os"
)

func AllFurnitureInDir(dir string) (paths,names []string) {
  filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
    if !info.IsDir() {
      if len(info.Name()) >= 5 && info.Name()[len(info.Name()) - 5 : ] == ".json" {
        var f Furniture
        err := base.LoadJson(path, &f)
        if err == nil {
          paths = append(paths, path)
          names = append(names, f.Name)
        }
      }
    }
    return nil
  })
  return
}

func LoadFurniture(path string) (*Furniture, error) {
  var f Furniture
  err := base.LoadJson(path, &f)
  if err != nil {
    return nil, err
  }
  texture_path := filepath.Clean(filepath.Join(path, f.Texture_path))
  f.texture_data = texture.LoadFromPath(texture_path)
  return &f, nil
}

type Furniture struct {
  // Name of the object - should be unique among all furniture
  Name string

  // Dimensions of the object in cells
  Dx,Dy int

  // Position of this object in board coordinates.  These values are only set
  // for objects that have been instantiated
  X,Y int

  // Path to the texture - relative to the location of this file
  Texture_path string

  // The texture itself
  texture_data *texture.Data
}

func (f *Furniture) Pos() (int, int) {
  return f.X, f.Y
}

func (f *Furniture) Dims() (int, int) {
  return f.Dx, f.Dy
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

func (f *Furniture) RenderDims(pos mathgl.Vec2, width float32) {
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

func (f *Furniture) Render(pos mathgl.Vec2, width float32) {
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
