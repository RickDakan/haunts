package game

import (
  "image"
  "strings"
  "path/filepath"
  "github.com/runningwild/glop/sprite"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game/status"
  "github.com/runningwild/haunts/house"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/mathgl"
  "github.com/runningwild/opengl/gl"
)

type Ai interface {
  Eval()
  Actions() <-chan Action
}

var ai_maker func(path string, ent *Entity) Ai

func SetAiMaker(f func(path string, ent *Entity) Ai) {
  ai_maker = f
}

func LoadAllEntities() {
  base.RemoveRegistry("entities")
  base.RegisterRegistry("entities", make(map[string]*entityDef))
  basedir := base.GetDataDir()
  base.RegisterAllObjectsInDir("entities", filepath.Join(basedir, "entities"), ".json", "json")
  base.RegisterAllObjectsInDir("entities", filepath.Join(basedir, "objects"), ".json", "json")
}

// Tries to place new_ent in the game at its current position.  Returns true
// on success, false otherwise.  If initial is true it will only place spawns
// in their appropriate spawn regions, as defined by spawn points.
func (g *Game) placeEntity(initial bool) bool {
  if g.new_ent == nil {
    return false
  }
  ix,iy := int(g.new_ent.X), int(g.new_ent.Y)
  idx,idy := g.new_ent.Dims()
  r, f := g.house.Floors[0].RoomAndFurnAtPos(ix, iy)
  if r == nil || f != nil { return false }
  for _,e := range g.Ents {
    x,y := e.Pos()
    dx,dy := e.Dims()
    r1 := image.Rect(x, y, x+dx, y+dy)
    r2 := image.Rect(ix, iy, ix+idx, iy+idy)
    if r1.Overlaps(r2) {
      return false
    }
  }

  // Check for spawn points, if it is an initial placement
  placeable := false
  if initial {
    haunt := g.new_ent.HauntEnt
    if haunt != nil {
      for _, sp := range g.house.Floors[0].Haunts {
        if sp == nil {
          continue
        }
        if !strings.HasPrefix(sp.Name, string(haunt.Level)) {
          continue
        }
        x, y := sp.Furniture().Pos()
        dx := x - ix
        if dx < 0 { dx = -dx }
        dy := y - iy
        if dy < 0 { dy = -dy }
        max := dx
        if dy > max { max = dy}
        if max > sp.Size {
          continue
        }
        placeable = true
        break
      }
    }
  } else {
    placeable = true
  }
  if !placeable {
    return false
  }

  g.Ents = append(g.Ents, g.new_ent)
  g.new_ent = nil
  return true
}

func MakeEntity(name string, g *Game) *Entity {
  ent := Entity{ Defname: name }
  base.GetObject("entities", &ent)
  for _,action_name := range ent.Action_names {
    ent.Actions = append(ent.Actions, MakeAction(action_name))
  }
  ent.Sprite.Load(ent.Sprite_path.String())

  if ent.Ai_path.String() != "" {
    ent.Ai = ai_maker(ent.Ai_path.String(), &ent)
  }

  if ent.Side() == SideHaunt || ent.Side() == SideExplorers {
    ent.los = &losData{}
    full_los := make([]bool, 256*256)
    ent.los.grid = make([][]bool, 256)
    for i := range ent.los.grid {
      ent.los.grid[i] = full_los[i * 256 : (i + 1) * 256]
    }
    stats := status.MakeInst(ent.Base)
    ent.Stats = &stats
  }

  ent.game = g
  return &ent
}

type spriteContainer struct {
  sp   *sprite.Sprite

  // If there is an error when loading the sprite it will be stored here
  err  error
}
func (sc *spriteContainer) Sprite() *sprite.Sprite {
  return sc.sp
}
func (sc *spriteContainer) Load(path string) {
  sc.sp, sc.err = sprite.LoadSprite(path)
}

// Allows the Ai system to signal to us under certain circumstance
type AiEvalSignal int
const (
  AiEvalCont  AiEvalSignal = iota
  AiEvalTerm
  AiEvalPause
)

type entityDef struct {
  Name string

  Sprite_path base.Path

  // Still frame of the sprite - not necessarily one of the individual frames,
  // but still usable for identifying it.  Should be the same dimensions as
  // any of the frames.
  Still    texture.Object  `registry:"autoload"`

  // Headshot of this character.  Should be square.
  Headshot texture.Object  `registry:"autoload"`

  // List of actions that this entity defaults to having
  Action_names []string

  // Path to the Ai that this entity should use if not player-controlled
  Ai_path base.Path

  Base status.Base

  ExplorerEnt *ExplorerEnt
  HauntEnt    *HauntEnt
  ObjectEnt   *ObjectEnt
}
func (ei *entityDef) Side() Side {
  types := 0
  if ei.ExplorerEnt != nil {
    types++
  }
  if ei.HauntEnt != nil {
    types++
  }
  if ei.ObjectEnt != nil {
    types++
  }
  if types != 1 {
    base.Error().Printf("Entity '%s' must specify exactly one ent type.", ei.Name)
    return SideNone
  }

  switch {
  case ei.ExplorerEnt != nil:
    return SideExplorers

  case ei.HauntEnt != nil:
    switch ei.HauntEnt.Level {
    case LevelMinion:
    case LevelMaster:
    case LevelServitor:
    default:
      base.Error().Printf("Entity '%s' speciied unknown level '%s'.", ei.Name, ei.HauntEnt.Level)
    }
    return SideHaunt

  case ei.ObjectEnt != nil:
    return SideObject
  }

  base.Error().Printf("This code should have been unreachable - offending entity: '%s'", ei.Name)
  return SideNone
}
func (ei *entityDef) Dims() (int,int) {
  return 2, 2
}

type HauntEnt struct {
  // If this entity is a Master, Cost indicates how many points it can spend
  // on Servitors, otherwise it indicates how many points a Master must pay to
  // include this entity in its army.
  Cost    int

  // If this entity is a Master this indicates how many points worth of
  // minions it begins the game with.  Not used for non-Masters.
  Minions int

  Level   EntLevel
}
type ExplorerEnt struct {}
type ObjectEnt struct {}

type EntLevel string
const(
  LevelMinion   EntLevel = "Minion"
  LevelServitor EntLevel = "Servitor"
  LevelMaster   EntLevel = "Master"
)

type Side int
const(
  SideNone      Side = iota
  SideExplorers
  SideHaunt
  SideObject
)

type losData struct {
  // All positions that can be seen by this entity are stored here.
  grid [][]bool

  // Floor coordinates of the last position los was determined from, so that
  // we don't need to recalculate it more than we need to as an ent is moving.
  x,y int

  // Range of vision - all true values in grid are contained within these
  // bounds.
  minx,miny,maxx,maxy int
}

type EntityInst struct {
  X,Y float64

  Sprite spriteContainer

  los *losData

  // The width that this entity's sprite was rendered at the last time it was
  // drawn.  User to determine what entity the cursor is over.
  last_render_width float32

  // Some methods may require being able to access other entities, so each
  // entity has a pointer to the game itself.
  game *Game

  // Actions that this entity currently has available to it for use.  This
  // may not be a bijection of Actions mentioned in entityDef.Action_names.
  Actions []Action

  Stats *status.Inst

  // Ai stuff - the channels cannot be gobbed, so they need to be remade when
  // loading an ent from a file
  Ai Ai
  // The ready flag is set to true at the start of every turn - this lets us
  // keep easy track of whether or not an entity's Ai has executed yet, since
  // an entity might not do anything else obvious, like run out of Ap.
  ai_status aiStatus

  // For inanimate objects - some of them need to be activated so we know when
  // the players can interact with them.
  Active bool
}
type aiStatus int
const (
  aiNone    aiStatus = iota
  aiReady
  aiRunning
  aiDone
)
func (e *Entity) Game() *Game {
  return e.game
}
func (e *Entity) HasLos(x,y int) bool {
  if e.los == nil {
    return false
  }
  if x < 0 || y < 0 || x >= len(e.los.grid) || y >= len(e.los.grid[0]) {
    return false
  }
  return e.los.grid[x][y]
}
func DiscretizePoint32(x,y float32) (int,int) {
  return DiscretizePoint64(float64(x), float64(y))
}
func DiscretizePoint64(x,y float64) (int,int) {
  x += 0.5
  y += 0.5
  if x < 0 {
    x -= 1
  }
  if y < 0 {
    y -= 1
  }
  return int(x), int(y)
}
func (ei *EntityInst) Pos() (int,int) {
  return DiscretizePoint64(ei.X, ei.Y)
}
func (ei *EntityInst) FPos() (float64,float64) {
  return ei.X, ei.Y
}

type Entity struct {
	Defname string
  *entityDef
  EntityInst
}

func (e *Entity) DrawReticle(viewer house.Viewer, ally,selected bool) {
  bx,by := e.FPos()
  wx,wy := viewer.BoardToWindow(float32(bx), float32(by))
  gl.Disable(gl.TEXTURE_2D)

  if ally {
    if selected {
      gl.Color4d(0, 1, 0, 0.8)
    } else {
      gl.Color4d(0, 1, 0, 0.3)
    }
  } else {
    if selected {
      gl.Color4d(1, 0, 0, 0.8)
    } else {
      gl.Color4d(1, 0, 0, 0.3)
    }
  }

  gl.Begin(gl.LINES)
    gl.Vertex2f(float32(wx) - e.last_render_width / 2, float32(wy))
    gl.Vertex2f(float32(wx) - e.last_render_width / 2, float32(wy) + 150 * e.last_render_width / 100)
    gl.Vertex2f(float32(wx) - e.last_render_width / 2, float32(wy) + 150 * e.last_render_width / 100)
    gl.Vertex2f(float32(wx) + e.last_render_width / 2, float32(wy) + 150 * e.last_render_width / 100)
    gl.Vertex2f(float32(wx) + e.last_render_width / 2, float32(wy) + 150 * e.last_render_width / 100)
    gl.Vertex2f(float32(wx) + e.last_render_width / 2, float32(wy))
    gl.Vertex2f(float32(wx) + e.last_render_width / 2, float32(wy))
    gl.Vertex2f(float32(wx) - e.last_render_width / 2, float32(wy))
  gl.End()
}

func (e *Entity) Render(pos mathgl.Vec2, width float32) {
  e.last_render_width = width
  gl.Enable(gl.TEXTURE_2D)
  if e.Sprite.sp != nil {
    tx,ty,tx2,ty2 := e.Sprite.sp.Bind()
    gl.Begin(gl.QUADS)
    gl.TexCoord2d(tx, -ty)
    gl.Vertex2f(pos.X - width/2, pos.Y)
    gl.TexCoord2d(tx, -ty2)
    gl.Vertex2f(pos.X - width/2, pos.Y + 150 * width / 100)
    gl.TexCoord2d(tx2, -ty2)
    gl.Vertex2f(pos.X + width/2, pos.Y + 150 * width / 100)
    gl.TexCoord2d(tx2, -ty)
    gl.Vertex2f(pos.X + width/2, pos.Y)
    gl.End()
  }
}

func facing(v mathgl.Vec2) int {
  fs := []mathgl.Vec2{
    mathgl.Vec2{ -1, -1 },
    mathgl.Vec2{ -4,  1 },
    mathgl.Vec2{  0,  1 },
    mathgl.Vec2{  1,  1 },
    mathgl.Vec2{  1,  0 },
    mathgl.Vec2{  1, -4 },
  }

  var max float32
  ret := 0
  for i := range fs {
    fs[i].Normalize()
    dot := fs[i].Dot(&v)
    if dot > max {
      max = dot
      ret = i
    }
  }
  return ret
}

func (e *Entity) TurnToFace(x,y int) {
  target := mathgl.Vec2{ float32(x), float32(y) }
  source := mathgl.Vec2{ float32(e.X), float32(e.Y) }
  var seg mathgl.Vec2
  seg.Assign(&target)
  seg.Subtract(&source)
  target_facing := facing(seg)
  f_diff := target_facing - e.Sprite.sp.Facing()
  if f_diff != 0 {
    f_diff = (f_diff + 6) % 6
    if f_diff > 3 {
      f_diff -= 6
    }
    for f_diff < 0 {
      e.Sprite.sp.Command("turn_left")
      f_diff++
    }
    for f_diff > 0 {
      e.Sprite.sp.Command("turn_right")
      f_diff--
    }
  }
}

// Advances ent up to dist towards the target cell.  Returns the distance
// traveled.
func (e *Entity) DoAdvance(dist float32, x,y int) float32 {
  if dist <= 0 {
    e.Sprite.sp.Command("stop")
    return 0
  }
  e.Sprite.sp.Command("move")

  source := mathgl.Vec2{ float32(e.X), float32(e.Y) }
  target := mathgl.Vec2{ float32(x), float32(y) }
  var seg mathgl.Vec2
  seg.Assign(&target)
  seg.Subtract(&source)
  e.TurnToFace(x, y)
  var traveled float32
  if seg.Length() > dist {
    seg.Scale(dist / seg.Length())
    traveled = dist
  } else {
    traveled = seg.Length()
  }
  seg.Add(&source)
  e.X = float64(seg.X)
  e.Y = float64(seg.Y)

  return dist - traveled
}

func (e *Entity) Think(dt int64) {
  if e.Sprite.sp != nil {
    e.Sprite.sp.Think(dt)
  }
}

func (e *Entity) OnRound() {
  if e.Stats != nil {
    e.Stats.OnRound()
    if e.Stats.HpCur() <= 0 {
      e.Sprite.Sprite().Command("defend")
      e.Sprite.Sprite().Command("killed")
    }
  }
}
