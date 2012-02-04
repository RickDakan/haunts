package actions

import (
  "glop/gin"
  "glop/gui"
  "glop/util/algorithm"
  "haunts/base"
  "haunts/game"
  "encoding/gob"
  "path/filepath"
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
  path [][2]int
}
type MoveDef struct {
  Name     string
}

func (a *Move) Readyable() bool {
  return true
}
func (a *Move) Prep(ent *game.Entity) bool {
  a.ent = ent
  return true
}
func (a *Move) HandleInput(group gui.EventGroup, g *game.Game) game.InputStatus {
  if found,event := group.FindEvent(gin.MouseLButton); found {
    src := g.ToVertex(a.ent.Pos())
    bx,by := g.GetViewer().WindowToBoard(event.Key.Cursor().Point())
    dst := g.ToVertex(int(bx), int(by))
    _,path := algorithm.Dijkstra(g, []int{src}, []int{dst})
    if len(path) <= 1 {
      return game.Consumed
    }
    a.path = algorithm.Map(path, [][2]int{}, func(a interface{}) interface{} {
      _,x,y := g.FromVertex(a.(int))
      return [2]int{ int(x), int(y) }
    }).([][2]int)
    return game.ConsumedAndBegin
  }
  return game.NotConsumed
}
func (a *Move) HandleOutput() {
}
func (a *Move) Cancel() {
  a.ent = nil
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

