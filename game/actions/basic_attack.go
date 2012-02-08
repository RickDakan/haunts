package actions

import (
  "glop/gui"
  "glop/gin"
  "haunts/base"
  "haunts/game/status"
  "haunts/game"
  "encoding/gob"
  "path/filepath"
  "gl"
)

func registerAttacks() map[string]func() game.Action {
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
  game.RegisterActionMakers(registerAttacks)
  gob.Register(&BasicAttack{})
}

// Basic Attacks are single target and instant, they are also readyable
type BasicAttack struct {
  Defname string
  *BasicAttackDef
  basicAttackInst
}
type BasicAttackDef struct {
  Name     string
  Kind     status.Kind
  Ap       int
  Strength int
  Range    int
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

  x,y := a.ent.Pos()
  for _,ent := range g.Ents {
    if ent == a.ent { continue }
    x2,y2 := ent.Pos()
    if dist(x, y, x2, y2) <= a.Range {
      a.targets = append(a.targets, ent)
    }
  }
  if len(a.targets) == 0 {
    a.ent = nil
    return false
  }
  return true
}
func (a *BasicAttack) HandleInput(group gui.EventGroup, g *game.Game) game.InputStatus {
  if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    cursor := event.Key.Cursor()
    fx, fy := g.GetViewer().WindowToBoard(cursor.Point())
    bx, by := int(fx), int(fy)
    for _,target := range a.targets {
      x,y := target.Pos()
      if bx == x && by == y {
        a.target = target
        return game.ConsumedAndBegin
      }
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
    a.ent.Sprite.Sprite().Command("melee")
    a.target.Sprite.Sprite().Command("defend")
    if game.DoAttack(a.ent, a.target, a.Strength, a.Kind) {
      a.target.Sprite.Sprite().Command("damaged")
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

