package actions

import (
  "glop/gin"
  "glop/gui"
  "glop/util/algorithm"
  "haunts/base"
  "haunts/game"
  "haunts/game/status"
  "encoding/gob"
  "path/filepath"
  "gl"
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
}

func (a *Move) Readyable() bool {
  return false
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
  _, path := algorithm.Dijkstra(ent.Game(), source_cell, dsts)
  if path == nil {
    return false
  }
  if len(path) <= 1 { // || !canPayForMove(a.Ent, a.Level.MakeBoardPosFromVertex(path[1])) {
    return false
  }
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
  }
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
func (a *Move) RenderOnFloor() {
  gl.Disable(gl.TEXTURE_2D)
  gl.Begin(gl.LINES)
  if a.cost <= a.ent.Stats.ApCur() {
    gl.Color4d(0.2, 0.5, 0.9, 0.8)
  } else {
    gl.Color4d(0.9, 0.5, 0.2, 0.8)
  }
  for i := 1; i < len(a.path); i++ {
    gl.Vertex2d(float64(a.path[i-1][0]) + 0.5, float64(a.path[i-1][1]) + 0.5)
    gl.Vertex2d(float64(a.path[i][0]) + 0.5, float64(a.path[i][1]) + 0.5)
  }
  gl.End()
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
    dist = a.ent.DoAdvance(float32(dt) / 200, a.path[0][0], a.path[0][1])
  }
  return game.InProgress
}
func (a *Move) Interrupt() bool {
  return true
}

