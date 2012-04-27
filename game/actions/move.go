package actions

import (
  "encoding/gob"
  "path/filepath"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/util/algorithm"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/house"
  "github.com/runningwild/haunts/game"
  "github.com/runningwild/haunts/game/status"
  "github.com/runningwild/haunts/texture"
  gl "github.com/chsc/gogl/gl21"
)

func registerMoves() map[string]func() game.Action {
  move_actions := make(map[string]*MoveDef)
  base.RemoveRegistry("actions-move_actions")
  base.RegisterRegistry("actions-move_actions", move_actions)
  base.RegisterAllObjectsInDir("actions-move_actions", filepath.Join(base.GetDataDir(), "actions", "movement"), ".json", "json")
  makers := make(map[string]func() game.Action)
  for name := range move_actions {
    cname := name
    makers[cname] = func() game.Action {
      a := Move{ Defname: cname }
      base.GetObject("actions-move_actions", &a)
      return &a
    }
  }
  return makers
}

var path_tex *house.LosTexture

func init() {
  game.RegisterActionMakers(registerMoves)
  gob.Register(&Move{})
}

type Move struct {
  Defname string
  *MoveDef

  ent *game.Entity

  // Destination that was used to generate path
  dst int

  // Whether or not we've trid to calculate a path to the currest dst vertex.
  // Since it's possible to not find a path we might end up with a nil path
  // even if we already tried that dst, and we don't want to have to keep
  // pathing if we don't need to.
  calculated bool

  path [][2]int
  cost int
}
type MoveDef struct {
  Name     string
  Texture  texture.Object
}
func (a *Move) AP() int {
  return a.cost
}
func (a *Move) String() string {
  return a.Name
}
func (a *Move) Icon() *texture.Object {
  return &a.Texture
}
func (a *Move) Readyable() bool {
  return false
}

func limitPath(g *game.Game, start int, path []int, max int) []int {
  total := 0
  for last := 0; last < len(path); last++ {
    adj,cost := g.Adjacent(start)
    for index := range adj {
      if adj[index] == path[last] {
        total += int(cost[index])
        if total > max {
          return path[0 : last]
        }
        start = adj[index]
        break
      }
    }
  }
  return path
}

// Usable by ais, tries to find a path that moves the entity to within dist of
// the specified location.  Returns true if possible, false otherwise.  If it
// returns true it also begins execution, so it should become the current
// action.
func (a *Move) AiMoveToWithin(ent *game.Entity, tx,ty,dist int) bool {
  a.ent = ent
  var dsts []int
  for x := tx - dist; x <= tx + dist; x++ {
    for y := ty - dist; y <= ty + dist; y++ {
      if x == tx && y == ty { continue }
      dsts = append(dsts, a.ent.Game().ToVertex(x, y))
    }
  }
  source_cell := []int{a.ent.Game().ToVertex(a.ent.Pos())}
  fcost, path := algorithm.Dijkstra(ent.Game(), source_cell, dsts)
  cost := int(fcost)
  if path == nil {
    return false
  }
  path = limitPath(ent.Game(), source_cell[0], path, ent.Stats.ApCur())
  if len(path) <= 1 { // || !canPayForMove(a.Ent, a.Level.MakeBoardPosFromVertex(path[1])) {
    return false
  }
  if cost > ent.Stats.ApCur() {
    cost = ent.Stats.ApCur()
  }
  ent.Stats.ApplyDamage(-cost, 0, status.Unspecified)
  vertex_to_boardpos := func(v interface{}) interface{} {
    _,x,y := a.ent.Game().FromVertex(v.(int))
    return [2]int{x,y}
  }
  a.path = algorithm.Map(path[1:], [][2]int{}, vertex_to_boardpos).([][2]int)
  return true
}

func (a *Move) findPath(g *game.Game, x,y int) {
  dst := g.ToVertex(x, y)
  if dst != a.dst || !a.calculated {
    a.dst = dst
    a.calculated = true
    src := g.ToVertex(a.ent.Pos())
    cost,path := algorithm.Dijkstra(g, []int{src}, []int{dst})
    if len(path) <= 1 {
      return
    }
    a.path = algorithm.Map(path, [][2]int{}, func(a interface{}) interface{} {
      _,x,y := g.FromVertex(a.(int))
      return [2]int{ int(x), int(y) }
    }).([][2]int)
    a.cost = int(cost)

    if path_tex != nil {
      pix := path_tex.Pix()
      for i := range pix {
        for j := range pix[i] {
          pix[i][j] = 0
        }
      }
      for _, v := range a.path {
        pix[v[1]][v[0]] = 255
      }
      path_tex.Remap()
    }
  }
}

func (a *Move) Preppable(ent *game.Entity, g *game.Game) bool {
  return true
}
func (a *Move) Prep(ent *game.Entity, g *game.Game) bool {
  a.ent = ent
  fx, fy := g.GetViewer().WindowToBoard(gin.In().GetCursor("Mouse").Point())
  a.findPath(g, int(fx), int(fy))
  return true
}
func (a *Move) HandleInput(group gui.EventGroup, g *game.Game) game.InputStatus {
  cursor := group.Events[0].Key.Cursor()
  if cursor != nil {
    fx, fy := g.GetViewer().WindowToBoard(cursor.Point())
    a.findPath(g, int(fx), int(fy))
  }
  if found,_ := group.FindEvent(gin.MouseLButton); found {
    if len(a.path) > 0 {
      if a.cost <= a.ent.Stats.ApCur() {
        a.ent.Stats.ApplyDamage(-a.cost, 0, status.Unspecified)
        a.cost = 0
        return game.ConsumedAndBegin
      }
      return game.Consumed
    } else {
      return game.NotConsumed
    }
  }
  return game.NotConsumed
}
func (a *Move) RenderOnFloor(room *house.Room) {
  if room == nil || a.ent == nil {
    return
  }
  if path_tex == nil {
    path_tex = house.MakeLosTexture()
  }
  path_tex.Remap()
  path_tex.Bind()
  if a.cost <= a.ent.Stats.ApCur() {
    gl.Color4ub(25, 255, 100, 255)
  } else {
    gl.Color4ub(255, 25, 25, 255)
  }
  texture.RenderAdvanced(float64(-room.X), float64(-room.Y), house.LosTextureSize, house.LosTextureSize, 3.1415926535, false)
}
func (a *Move) Cancel() {
  a.ent = nil
  a.path = nil
  a.calculated = false
}
func (a *Move) Maintain(dt int64) game.MaintenanceStatus {
  // Do stuff
  dist := a.ent.DoAdvance(float32(dt) / 200, a.path[0][0], a.path[0][1])
  for dist > 0 {
    if len(a.path) == 1 {
      a.ent.DoAdvance(0,0,0)
      a.ent = nil
      return game.Complete
    }
    a.path = a.path[1:]
    dist = a.ent.DoAdvance(dist, a.path[0][0], a.path[0][1])
  }
  return game.InProgress
}
func (a *Move) Interrupt() bool {
  return true
}

