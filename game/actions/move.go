package actions

import (
  "math"
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

  // Ap remaining before the ability was used
  threshold int
}
type MoveDef struct {
  Name     string
  Texture  texture.Object
}

type moveExec struct {
  game.BasicActionExec
  Dst int
}
func init() {
  gob.Register(moveExec{})
}

func (a *Move) AP() int {
  return a.cost
}
func (a *Move) Pos() (int, int) {
  return 0, 0
}
func (a *Move) Dims() (int, int) {
  return house.LosTextureSize, house.LosTextureSize
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
func (a *Move) AiMoveToWithin(ent *game.Entity, tx,ty,dist int) game.ActionExec {
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
    return nil
  }
  path = limitPath(ent.Game(), source_cell[0], path, ent.Stats.ApCur())
  if len(path) <= 1 { // || !canPayForMove(a.Ent, a.Level.MakeBoardPosFromVertex(path[1])) {
    return nil
  }
  var exec moveExec
  exec.SetBasicData(a.ent, a)
  exec.Dst = path[len(path)-1]
  return exec
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
      current := 0.0
      for i := 1; i < len(a.path); i++ {
        src := g.ToVertex(a.path[i-1][0], a.path[i-1][1])
        dst := g.ToVertex(a.path[i][0], a.path[i][1])
        v, cost := g.Adjacent(src)
        for j := range v {
          if v[j] == dst {
            current += cost[j]
            break
          }
        }
        pix[a.path[i][1]][a.path[i][0]] += byte(current)
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
  a.threshold = a.ent.Stats.ApCur()
  return true
}
func (a *Move) HandleInput(group gui.EventGroup, g *game.Game) (bool, game.ActionExec) {
  cursor := group.Events[0].Key.Cursor()
  if cursor != nil {
    fx, fy := g.GetViewer().WindowToBoard(cursor.Point())
    a.findPath(g, int(fx), int(fy))
  }
  if found,_ := group.FindEvent(gin.MouseLButton); found {
    if len(a.path) > 0 {
      if a.cost <= a.ent.Stats.ApCur() {
        var exec moveExec
        exec.SetBasicData(a.ent, a)
        exec.Dst = a.dst
        return true, exec
      }
      return true, nil
    } else {
      return false, nil
    }
  }
  return false, nil
}
func (a *Move) RenderOnFloor() {
  if a.ent == nil {
    return
  }
  if path_tex == nil {
    path_tex = house.MakeLosTexture()
  }
  path_tex.Remap()
  path_tex.Bind()
  gl.Color4ub(255, 255, 255, 128)
  base.EnableShader("path")
  base.SetUniformF("path", "threshold", float32(a.threshold) / 255)
  base.SetUniformF("path", "size", house.LosTextureSize)
  texture.RenderAdvanced(0, 0, house.LosTextureSize, house.LosTextureSize, 3.1415926535, false)
  base.EnableShader("")
}
func (a *Move) Cancel() {
  a.ent = nil
  a.path = nil
  a.calculated = false
}
func (a *Move) Maintain(dt int64, ae game.ActionExec) game.MaintenanceStatus {
  if ae != nil {
    exec := ae.(moveExec)
    _, x, y := a.ent.Game().FromVertex(exec.Dst)
    a.findPath(a.ent.Game(), x, y)
    a.ent.Stats.ApplyDamage(-a.cost, 0, status.Unspecified)
  }
  // Do stuff
  factor := float32(math.Pow(2, a.ent.Walking_speed))
  dist := a.ent.DoAdvance(factor * float32(dt) / 200, a.path[0][0], a.path[0][1])
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

