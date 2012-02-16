package actions

import (
  "encoding/gob"
  "path/filepath"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  "github.com/runningwild/haunts/game/status"
  "github.com/runningwild/opengl/gl"
)

func registerBasicAttacks() map[string]func() game.Action {
  attack_actions := make(map[string]*BasicAttackDef)
  base.RemoveRegistry("actions-attack_actions")
  base.RegisterRegistry("actions-attack_actions", attack_actions)
  base.RegisterAllObjectsInDir("actions-attack_actions", filepath.Join(base.GetDataDir(), "actions", "basic_attacks"), ".json", "json")
  makers := make(map[string]func() game.Action)
  for name := range attack_actions {
    cname := name
    makers[cname] = func() game.Action {
      a := BasicAttack{ Defname: cname }
      base.GetObject("actions-attack_actions", &a)
      return &a
    }
  }
  return makers
}

func init() {
  game.RegisterActionMakers(registerBasicAttacks)
  gob.Register(&BasicAttack{})
}

// Basic Attacks are single target and instant, they are also readyable
type BasicAttack struct {
  Defname string
  *BasicAttackDef
  basicAttackInst
}
type BasicAttackDef struct {
  Name       string
  Kind       status.Kind
  Ap         int
  Strength   int
  Range      int
  Damage     int
  Animation  string
  Conditions []string
}
type basicAttackInst struct {
  ent *game.Entity

  // Potential targets
  targets []*game.Entity

  // The selected target for the attack
  target *game.Entity
}
func dist(x,y,x2,y2 int) int {
  dx := x - x2
  if dx < 0 { dx = -dx }
  dy := y - y2
  if dy < 0 { dy = -dy }
  if dx > dy {
    return dx
  }
  return dy
}
func (a *BasicAttack) Readyable() bool {
  return true
}
func (a *BasicAttack) Prep(ent *game.Entity, g *game.Game) bool {
  a.ent = ent
  a.targets = nil

  if a.ent.Stats.ApCur() < a.Ap {
    return false
  }

  x,y := a.ent.Pos()
  for _,ent := range g.Ents {
    if ent == a.ent { continue }
    x2,y2 := ent.Pos()
    if dist(x, y, x2, y2) <= a.Range && a.ent.HasLos(x2, y2) && ent.Stats.HpCur() > 0 {
      a.targets = append(a.targets, ent)
    }
  }
  if len(a.targets) == 0 {
    a.ent = nil
    return false
  }
  return true
}
func (a *BasicAttack) AiAttackTarget(ent *game.Entity, target *game.Entity) bool {
  if ent.Side == target.Side { return false }
  if ent.Stats.ApCur() < a.Ap { return false }
  x,y := ent.Pos()
  x2,y2 := target.Pos()
  if dist(x,y,x2,y2) > a.Range { return false }
  a.ent = ent
  a.target = target
  a.ent.Stats.ApplyDamage(-a.Ap, 0, status.Unspecified)
  return true
}
func (a *BasicAttack) HandleInput(group gui.EventGroup, g *game.Game) game.InputStatus {
  target := g.HoveredEnt()
  if target == nil { return game.NotConsumed }
  if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    if a.ent.Stats.ApCur() >= a.Ap && target.Stats.HpCur() > 0 && a.ent.HasLos(target.Pos()) {
      a.target = target
      a.ent.Stats.ApplyDamage(-a.Ap, 0, status.Unspecified)
      return game.ConsumedAndBegin
    }
    return game.Consumed
  }
  return game.NotConsumed
}
func (a *BasicAttack) RenderOnFloor() {
  gl.Disable(gl.TEXTURE_2D)
  gl.Begin(gl.QUADS)
  gl.Color4d(1.0, 0.2, 0.2, 0.8)
  for _,ent := range a.targets {
    ix,iy := ent.Pos()
    x := float64(ix)
    y := float64(iy)
    gl.Vertex2d(x + 0, y + 0)
    gl.Vertex2d(x + 0, y + 1)
    gl.Vertex2d(x + 1, y + 1)
    gl.Vertex2d(x + 1, y + 0)
  }
  gl.End()
}
func (a *BasicAttack) Cancel() {
  a.basicAttackInst = basicAttackInst{}
}
func (a *BasicAttack) Maintain(dt int64) game.MaintenanceStatus {
  if a.ent.Sprite.Sprite().State() == "ready" && a.target.Sprite.Sprite().State() == "ready" {
    a.target.TurnToFace(a.ent.Pos())
    a.ent.TurnToFace(a.target.Pos())
    a.ent.Sprite.Sprite().Command(a.Animation)
    a.target.Sprite.Sprite().Command("defend")
    if game.DoAttack(a.ent, a.target, a.Strength, a.Kind) {
      for _,name := range a.Conditions {
        a.target.Stats.ApplyCondition(status.MakeCondition(name))
      }
      a.target.Stats.ApplyDamage(0, -a.Damage, a.Kind)
      if a.target.Stats.HpCur() <= 0 {
        a.target.Sprite.Sprite().Command("killed")
      } else {
        a.target.Sprite.Sprite().Command("damaged")
      }
    } else {
      a.target.Sprite.Sprite().Command("undamaged")
    }
    return game.Complete
  }
  return game.InProgress
}
func (a *BasicAttack) Interrupt() bool {
  return true
}

