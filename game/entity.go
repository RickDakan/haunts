package game

import (
  // "fmt"
  "glop/sprite"
  "haunts/base"
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
  return &ent
}

type spriteContainer struct {
  Path string          `registry:"path"`
  sp   *sprite.Sprite

  // If there is an error when loading the sprite it will be stored here
  err  error
}
func (sc *spriteContainer) Load() {
  sc.sp, sc.err = sprite.LoadSprite(sc.Path)
}

type entityDef struct {
  Name string
  Sprite spriteContainer  `registry:"autoload"`
}
func (ei *entityDef) Dims() (int,int) {
  return 1, 1
}

type EntityInst struct {
  X,Y float64

  // If the entity is currently moving along a path it will be stored here.
  // Path[0] is the cell that the entity is currently moving directly towards.
  Path [][2]int
}
func (ei *EntityInst) Pos() (int,int) {
  x := ei.X + 0.5
  y := ei.Y + 0.5
  if x < 0 {
    x -= 1
  }
  if y < 0 {
    y -= 1
  }
  return int(x), int(y)
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
    mathgl.Vec2{ -2, -1 },
    mathgl.Vec2{ -2, 1 },
    mathgl.Vec2{ 1, 1 },
    mathgl.Vec2{ 2, 1 },
    mathgl.Vec2{ 2, -1 },
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

func (e *Entity) advance(dist float32) {
  if len(e.Path) == 0 {
    e.Sprite.sp.Command("stop")
    return
  } else {
    e.Sprite.sp.Command("move")
  }
  if dist <= 0 { return }
  target := mathgl.Vec2{ float32(e.Path[0][0]), float32(e.Path[0][1]) }
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
      println("left")
      f_diff++
    }
    for f_diff > 0 {
      e.Sprite.sp.Command("turn_right")
      f_diff--
      println("right")
    }
  }
  var traveled float32
  if seg.Length() > dist {
    seg.Scale(dist / seg.Length())
    traveled = dist
  } else {
    traveled = seg.Length()
    e.Path = e.Path[1:]
  }
  seg.Add(&source)
  e.X = float64(seg.X)
  e.Y = float64(seg.Y)
  e.advance(dist - traveled)
}

func (e *Entity) Think(dt int64) {
  if e.Sprite.sp != nil {
    e.Sprite.sp.Think(dt)
    e.advance(float32(dt) / 200)
  }
}
