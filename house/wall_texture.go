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
  // Position of the texture in floor coordinates.  If these coordinates exceed
  // either the dx or dy of the room, then this texture will be drawn, at least
  // partially, on the wall.  The coordinates should not both exceed the
  // dimensions of the room.
  X,Y float32
  Rot float32
}

func (wt *WallTexture) Render() {
  dx2 := float32(wt.texture_data.Dx) / 100 / 2
  dy2 := float32(wt.texture_data.Dy) / 100 / 2
  gl.Enable(gl.TEXTURE_2D)
  wt.texture_data.Bind()

  var rot mathgl.Mat3
  rot.RotationZ(wt.Rot)

  ll := mathgl.Vec2{ - dx2, - dy2 }
  ul := mathgl.Vec2{ - dx2, + dy2 }
  ur := mathgl.Vec2{ + dx2, + dy2 }
  lr := mathgl.Vec2{ + dx2, - dy2 }

  ll.Transform(&rot)
  ul.Transform(&rot)
  ur.Transform(&rot)
  lr.Transform(&rot)

  gl.Color4f(1, 1, 1, 1)
  gl.Begin(gl.QUADS)
  gl.TexCoord2i(0, 0)
  gl.Vertex2f(wt.X + ll.X, wt.Y + ll.Y)
  gl.TexCoord2i(0, -1)
  gl.Vertex2f(wt.X + ul.X, wt.Y + ul.Y)
  gl.TexCoord2i(-1, -1)
  gl.Vertex2f(wt.X + ur.X, wt.Y + ur.Y)
  gl.TexCoord2i(-1, 0)
  gl.Vertex2f(wt.X + lr.X, wt.Y + lr.Y)
  gl.End()
}
