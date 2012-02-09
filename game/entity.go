package game

import (
  "glop/sprite"
  "haunts/base"
  "haunts/game/status"
  "github.com/arbaal/mathgl"
  "gl"
)

func LoadAllEntitiesInDir(dir string) {
  base.RemoveRegistry("entities")
  base.RegisterRegistry("entities", make(map[string]*entityDef))
  base.RegisterAllObjectsInDir("entities", dir, ".json", "json")
}

func MakeEntity(name string) *Entity {
  ent := Entity{ Defname: name }
  base.GetObject("entities", &ent)
  for _,action_name := range ent.Action_names {
    ent.Actions = append(ent.Actions, MakeAction(action_name))
  }
  return &ent
}

type spriteContainer struct {
  Path base.Path
  sp   *sprite.Sprite

  // If there is an error when loading the sprite it will be stored here
  err  error
}
func (sc *spriteContainer) Sprite() *sprite.Sprite {
  return sc.sp
}
func (sc *spriteContainer) Load() {
  sc.sp, sc.err = sprite.LoadSprite(sc.Path.String())
}

type entityDef struct {
  Name string
  Sprite spriteContainer  `registry:"autoload"`

  // List of actions that this entity defaults to having
  Action_names []string

  // Contains both base and current stats
  Stats status.Inst
}
func (ei *entityDef) Dims() (int,int) {
  return 1, 1
}

type EntityInst struct {
  X,Y float64

  // All positions that can be seen by this entity are stored here.
  los map[[2]int]bool

  // Floor coordinates of the last position los was determined from, so that
  // we don't need to recalculate it more than we need to as an ent is moving.
  losx,losy int

  // Actions that this entity currently has available to it for use.  This
  // may not be a bijection of Actions mentioned in entityDef.Action_names.
  Actions []Action

  Status status.Inst
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

func (e *Entity) Render(pos mathgl.Vec2, width float32) {
  if e.Sprite.sp != nil {
    tx,ty,tx2,ty2 := e.Sprite.sp.Bind()
    gl.Color4d(1, 1, 1, 1)
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
  e.Stats.OnRound()
}
