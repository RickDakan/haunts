package actions

import (
  "encoding/gob"
  "path/filepath"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/util/algorithm"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/haunts/game/status"
)

func registerInteracts() map[string]func() game.Action {
  interact_actions := make(map[string]*InteractDef)
  base.RemoveRegistry("actions-interact_actions")
  base.RegisterRegistry("actions-interact_actions", interact_actions)
  base.RegisterAllObjectsInDir("actions-interact_actions", filepath.Join(base.GetDataDir(), "actions", "interacts"), ".json", "json")
  makers := make(map[string]func() game.Action)
  for name := range interact_actions {
    cname := name
    makers[cname] = func() game.Action {
      a := Interact{ Defname: cname }
      base.GetObject("actions-interact_actions", &a)
      return &a
    }
  }
  return makers
}

func init() {
  game.RegisterActionMakers(registerInteracts)
  gob.Register(&Interact{})
}

type Interact struct {
  Defname string
  *InteractDef
  interactInst
}
type InteractDef struct {
  Name         string  // "Relic", "Mystery", or "Cleanse"
  Display_name string  // The string actually displayed to the user
  Ap           int
  Range        int
  Animation    string
  Texture      texture.Object
}
type interactInst struct {
  ent *game.Entity

  // Potential targets
  targets []*game.Entity

  // The selected target for the attack
  target *game.Entity
}
type interactExec struct {
  game.BasicActionExec
  Target game.EntityId
}
func init() {
  gob.Register(interactExec{})
}
func (a *Interact) AP() int {
  return a.Ap
}
func (a *Interact) Pos() (int, int) {
  return 0, 0
}
func (a *Interact) Dims() (int, int) {
  return 0, 0
}
func (a *Interact) String() string {
  return a.Display_name
}
func (a *Interact) Icon() *texture.Object {
  return &a.Texture
}
func (a *Interact) Readyable() bool {
  return false
}
func distBetweenEnts(e1, e2 *game.Entity) int {
  x1,y1 := e1.Pos()
  dx1,dy1 := e1.Dims()
  x2,y2 := e2.Pos()
  dx2,dy2 := e2.Dims()

  var xdist int
  switch {
  case x1 >= x2 + dx2:
    xdist = x1 - (x2 + dx2)
  case x2 >= x1 + dx1:
    xdist = x2 - (x1 + dx1)
  default:
    xdist = 0
  }

  var ydist int
  switch {
  case y1 >= y2 + dy2:
    ydist = y1 - (y2 + dy2)
  case y2 >= y1 + dy1:
    ydist = y2 - (y1 + dy1)
  default:
    ydist = 0
  }

  if xdist > ydist {
    return xdist
  }
  return ydist
}
func (a *Interact) findTargets(ent *game.Entity, g *game.Game) []*game.Entity {
  var targets []*game.Entity
  for _,e := range g.Ents {
    x,y := e.Pos()
    dx,dy := e.Dims()
    if e == ent { continue }
    if e.ObjectEnt == nil { continue }
    if e.ObjectEnt.Goal != game.ObjectGoal(a.Name) { continue }
    if distBetweenEnts(e, ent) > a.Range { continue }
    if !ent.HasLos(x, y, dx, dy) { continue }

    // Make sure it's still active:
    active := false
    switch a.Name {
    case string(game.GoalCleanse):
      for i := range g.Active_cleanses {
        if g.Active_cleanses[i] == e {
          active = true
          break
        }
      }

    case string(game.GoalRelic):
      active = (e == g.Active_relic)

    case string(game.GoalMystery):
    }
    if !active { continue }

    targets = append(targets, e)
  }
  return targets
}
func (a *Interact) Preppable(ent *game.Entity, g *game.Game) bool {
  if a.Ap > ent.Stats.ApCur() {
    return false
  }
  a.targets = a.findTargets(ent, g)
  return len(a.targets) > 0
}
func (a *Interact) Prep(ent *game.Entity, g *game.Game) bool {
  if a.Preppable(ent, g) {
    a.ent = ent
    return true
  }
  return false
}
func (a *Interact) HandleInput(group gui.EventGroup, g *game.Game) (bool, game.ActionExec) {
  target := g.HoveredEnt()
  if target == nil { return false, nil }
  if found, event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    for i := range a.targets {
      if a.targets[i] == target {
        // switch a.Name {
        // case string(game.GoalCleanse):
        //   g.Active_cleanses = algorithm.Choose(g.Active_cleanses, func(a interface{}) bool {
        //     return a.(*game.Entity) != target
        //   }).([]*game.Entity)
    
        // case string(game.GoalRelic):
        //   g.Active_relic = nil
    
        // case string(game.GoalMystery):
        // }
        // a.ent.Stats.ApplyDamage(-a.Ap, 0, status.Unspecified)
        var exec interactExec
        exec.SetBasicData(a.ent, a)
        exec.Target = target.Id
        return true, exec
      }
    }
    return true, nil
  }
  return false, nil
}
func (a *Interact) RenderOnFloor() {
}
func (a *Interact) Cancel() {
  a.interactInst = interactInst{}
}
func (a *Interact) Maintain(dt int64, ae game.ActionExec) game.MaintenanceStatus {
  if ae != nil {
    exec := ae.(interactExec)
    g := a.ent.Game()
    target := g.EntityById(exec.Target)
    switch a.Name {
    case string(game.GoalCleanse):
      g.Active_cleanses = algorithm.Choose(g.Active_cleanses, func(a interface{}) bool {
        return a.(*game.Entity) != target
      }).([]*game.Entity)

    case string(game.GoalRelic):
      g.Active_relic = nil

    case string(game.GoalMystery):
    }
    a.ent.Stats.ApplyDamage(-a.Ap, 0, status.Unspecified)
    target.Sprite().Command("inspect")
  }
  return game.Complete
}
func (a *Interact) Interrupt() bool {
  return true
}

