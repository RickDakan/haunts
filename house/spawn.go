package house

import (
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/mathgl"
  gl "github.com/chsc/gogl/gl21"
)

type SpawnType int
const(
  SpawnRelic     SpawnType = iota
  SpawnExit
  SpawnExplorers
  SpawnHaunts
  SpawnClue
  SpawnCleanse
)

func MakeSpawnPoint(name string) *SpawnPoint {
  r := SpawnPoint{ Defname: name }
  base.GetObject("spawns", &r)
  return &r
}

func GetAllSpawnPointNames() []string {
  return base.GetAllNamesInRegistry("spawns")
}

func LoadAllSpawnPointsInDir(dir string) {
  base.RemoveRegistry("spawns")
  base.RegisterRegistry("spawns", make(map[string]*SpawnPointDef))
  base.RegisterAllObjectsInDir("spawns", dir, ".json", "json")
}

type HauntPoint struct {
  // Whether or not each type of haunt can spawn there
  Minions, Servitors, Masters bool
}
type ExplorerPoint struct {}
type CluePoint struct {}
type CleansePoint struct {}
type ExitPoint struct {}
type RelicPoint struct {}

type SpawnPointDef struct {
  Name string

  // Exactly one of the SpawnPoint types will be non-nil
  Haunt     *HauntPoint
  Explorer  *ExplorerPoint
  Clue      *CluePoint
  Cleanse   *CleansePoint
  Exit      *ExitPoint
  Relic     *RelicPoint
}
func (sp *SpawnPointDef) Type() SpawnType {
  count := 0
  if sp.Haunt != nil { count++ }
  if sp.Explorer != nil { count++ }
  if sp.Clue != nil { count++ }
  if sp.Cleanse != nil { count++ }
  if sp.Exit != nil { count++ }
  if sp.Relic != nil { count++ }
  if count > 1 {
    // This error will keep repeating - oh well
    base.Error().Printf("SpawnPointDef '%s' specified more than one spawn type", sp.Name)
  }
  switch {
  case sp.Haunt != nil:
    return SpawnHaunts
  case sp.Explorer != nil:
    return SpawnExplorers
  case sp.Clue != nil:
    return SpawnClue
  case sp.Cleanse != nil:
    return SpawnCleanse
  case sp.Exit != nil:
    return SpawnClue
  case sp.Relic != nil:
    return SpawnRelic
  }
  base.Error().Printf("SpawnPointDef '%s' didn't specify a spawn type", sp.Name)
  // Setting it to something so that error doesn't repeat again and again
  sp.Clue = &CluePoint{}
  return SpawnClue
}

type SpawnPoint struct {
  Defname string
  *SpawnPointDef
  Dx,Dy int
  X,Y   int
  temporary, invalid bool
}
func (sp *SpawnPoint) Dims() (int,int) {
  return sp.Dx, sp.Dy
}
func (sp *SpawnPoint) Pos() (int,int) {
  return sp.X, sp.Y
}
func (sp *SpawnPoint) FPos() (float64,float64) {
  return float64(sp.X), float64(sp.Y)
}
func (sp *SpawnPoint) Color() (r,g,b,a byte) {
  return 255, 255, 255, 255
}
func (sp *SpawnPoint) Render(pos mathgl.Vec2, width float32) {
  gl.Disable(gl.TEXTURE_2D)
  gl.Color4d(1, 1, 1, 0.1)
  gl.Begin(gl.QUADS)
    gl.Vertex2f(pos.X-width/2, pos.Y)
    gl.Vertex2f(pos.X-width/2, pos.Y+width)
    gl.Vertex2f(pos.X+width/2, pos.Y+width)
    gl.Vertex2f(pos.X+width/2, pos.Y)
  gl.End()
}
func (sp *SpawnPoint) RenderOnFloor() {
  var rgba [4]float64
  gl.GetDoublev(gl.CURRENT_COLOR, &rgba[0])
  gl.PushAttrib(gl.CURRENT_BIT)
  gl.Disable(gl.TEXTURE_2D)
  switch sp.Type() {
  case SpawnRelic:
    gl.Color4ub(255, 0, 255, byte(255 * rgba[3]))

  case SpawnClue:
    gl.Color4ub(0, 0, 255, byte(255 * rgba[3]))

  case SpawnCleanse:
    gl.Color4ub(255, 255, 255, byte(255 * rgba[3]))

  case SpawnExplorers:
    gl.Color4ub(0, 255, 0, byte(255 * rgba[3]))

  case SpawnHaunts:
    gl.Color4ub(255, 0, 0, byte(255 * rgba[3]))

  case SpawnExit:
    gl.Color4ub(0, 255, 255, byte(255 * rgba[3]))

  default:
    gl.Color4ub(0, 0, 0, byte(255 * rgba[3]))
  }
  base.EnableShader("box")
  base.SetUniformF("box", "dx", float32(sp.Dx))
  base.SetUniformF("box", "dy", float32(sp.Dy))
  if !sp.temporary {
    base.SetUniformI("box", "temp_invalid", 0)
  } else if !sp.invalid {
    base.SetUniformI("box", "temp_invalid", 1)
  } else {
    base.SetUniformI("box", "temp_invalid", 2)
  }
  (&texture.Object{}).Data().Render(float64(sp.X), float64(sp.Y), float64(sp.Dx), float64(sp.Dy))
  base.EnableShader("")
  gl.PopAttrib()
}
