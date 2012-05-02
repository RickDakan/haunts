package house

import (
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
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
