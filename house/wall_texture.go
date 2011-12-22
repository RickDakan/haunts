package house

import (
  "haunts/base"
  "haunts/texture"
  "gl"
  "path/filepath"
  "os"
  "sort"
)

var (
  wall_texture_registry map[string]*wallTextureDef
)

func init() {
  wall_texture_registry = make(map[string]*wallTextureDef)
}

func MakeWallTexture(name string) *WallTexture {
  wt := WallTexture{ Defname: name }
  wt.Load()
  return &wt
}

func GetAllWallTextureNames() []string {
  var names []string
  for name := range wall_texture_registry {
    names = append(names, name)
  }
  sort.Strings(names)
  return names
}

func LoadAllWallTexturesInDir(dir string) {
  filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
    if !info.IsDir() {
      if len(info.Name()) >= 5 && info.Name()[len(info.Name()) - 5 : ] == ".json" {
        var wt wallTextureDef
        err := base.LoadJson(path, &wt)
        if err == nil {
          wt.abs_texture_path = filepath.Clean(filepath.Join(path, wt.Texture_path))
          wall_texture_registry[wt.Name] = &wt
        }
      }
    }
    return nil
  })
}

func (wt *WallTexture) Load() {
  wt.wallTextureDef = wall_texture_registry[wt.Defname]
  if wt.wallTextureDef.texture_data == nil {
    wt.wallTextureDef.texture_data = texture.LoadFromPath(wt.wallTextureDef.abs_texture_path)
  }
}

type WallTexture struct {
  Defname string
  *wallTextureDef
  WallTextureInst
}

type wallTextureDef struct {
  // Name of this texture as it appears in the editor, should be unique among
  // all WallTextures
  Name string

  // Path to the texture - relative to the location of this file
  Texture_path string

  // Absolute path to texture
  abs_texture_path string

  // The texture itself
  texture_data *texture.Data
}

type WallTextureInst struct {
  // Position of the center of this texture along the length of the wall, 0
  // being the left-most edge of the wall, 1 being the right-most.
  Pos float32

  // Height, in units of cells, of the center of this texture above the base
  // of the wall.
  Height float32
}

func (wt *WallTexture) Render(dx,dy int) {
  dx2 := float32(wt.texture_data.Dx) / 100 / 2
  dz2 := float32(wt.texture_data.Dy) / 100 / 2
  gl.Enable(gl.TEXTURE_2D)
  wt.texture_data.Bind()
  wpos := float32(dx + dy) * wt.Pos
  gl.Begin(gl.QUADS)
    if wpos + dx2 < float32(dx) {
      gl.TexCoord2f(-1, 0)
      gl.Vertex3f(wpos - dx2, float32(dy), -wt.Height - dz2)
      gl.TexCoord2f(-1, 1)
      gl.Vertex3f(wpos - dx2, float32(dy), -wt.Height + dz2)
      gl.TexCoord2f(0, 1)
      gl.Vertex3f(wpos + dx2, float32(dy), -wt.Height + dz2)
      gl.TexCoord2f(0, 0)
      gl.Vertex3f(wpos + dx2, float32(dy), -wt.Height - dz2)
    } else if wpos - dx2 > float32(dx) {
      gl.TexCoord2f(1, 0)
      gl.Vertex3f(float32(dx), float32(dy + dx) - wpos - dx2, -wt.Height - dz2)
      gl.TexCoord2f(1, 1)
      gl.Vertex3f(float32(dx), float32(dy + dx) - wpos - dx2, -wt.Height + dz2)
      gl.TexCoord2f(0, 1)
      gl.Vertex3f(float32(dx), float32(dy + dx) - wpos + dx2, -wt.Height + dz2)
      gl.TexCoord2f(0, 0)
      gl.Vertex3f(float32(dx), float32(dy + dx) - wpos + dx2, -wt.Height - dz2)
    } else {
      // It spans a corner, so we need to draw it as two separate quads
//      var corner float32
      var corner float32
      left := wpos - dx2
      right := float32(dx + dy) - wpos - dx2
      corner = 1 - (float32(dx) - left) / (wpos + dx2 - left)
      gl.TexCoord2f(-1, 0)
      gl.Vertex3f(left, float32(dy), -wt.Height - dz2)
      gl.TexCoord2f(-1, 1)
      gl.Vertex3f(left, float32(dy), -wt.Height + dz2)
      gl.TexCoord2f(-corner, 1)
      gl.Vertex3f(float32(dx), float32(dy), -wt.Height + dz2)
      gl.TexCoord2f(-corner, 0)
      gl.Vertex3f(float32(dx), float32(dy), -wt.Height - dz2)

      gl.TexCoord2f(-corner, 0)
      gl.Vertex3f(float32(dx), float32(dx), -wt.Height - dz2)
      gl.TexCoord2f(-corner, 1)
      gl.Vertex3f(float32(dx), float32(dx), -wt.Height + dz2)
      gl.TexCoord2f(0, 1)
      gl.Vertex3f(float32(dx), right, -wt.Height + dz2)
      gl.TexCoord2f(0, 0)
      gl.Vertex3f(float32(dx), right, -wt.Height - dz2)

    }
  gl.End()
}
