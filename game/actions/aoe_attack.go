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

func registerAoeAttacks() map[string]func() game.Action {
  aoe_actions := make(map[string]*AoeAttackDef)
  base.RemoveRegistry("actions-aoe_actions")
  base.RegisterRegistry("actions-aoe_actions", aoe_actions)
  base.RegisterAllObjectsInDir("actions-aoe_actions", filepath.Join(base.GetDataDir(), "actions", "aoe_attacks"), ".json", "json")
  makers := make(map[string]func() game.Action)
  for name := range aoe_actions {
    cname := name
    makers[cname] = func() game.Action {
      a := AoeAttack{ Defname: cname }
      base.GetObject("actions-aoe_actions", &a)
      return &a
    }
  }
  return makers
}

func init() {
  game.RegisterActionMakers(registerAoeAttacks)
  gob.Register(&AoeAttack{})
}

// Aoe Attacks are untargeted and instant, they are also readyable
type AoeAttack struct {
  Defname string
  *AoeAttackDef
  aoeAttackInst
}
type AoeAttackDef struct {
  Name       string
  Kind       status.Kind
  Ap         int
  Strength   int
  Range      int
  Diameter   int
  Damage     int
  Animation  string
  Conditions []string
}
type aoeAttackInst struct {
  ent *game.Entity

  // position of the target cell - in the case of an even diameter this will be
  // the lower-left of the center
  tx,ty int

  // All entities in the blast radius - could include the acting entity
  targets []*game.Entity
}
func (a *AoeAttack) Readyable() bool {
  return true
}
func (a *AoeAttack) Prep(ent *game.Entity, g *game.Game) bool {
  a.ent = ent

  if a.ent.Stats.ApCur() < a.Ap {
    return false
  }

  bx,by := g.GetViewer().WindowToBoard(gin.In().GetCursor("Mouse").Point())
  a.tx = int(bx)
  a.ty = int(by)
  return true
}
func (a *AoeAttack) HandleInput(group gui.EventGroup, g *game.Game) game.InputStatus {
  cursor := group.Events[0].Key.Cursor()
  if cursor != nil && cursor.Name() == "Mouse" {
    bx,by := g.GetViewer().WindowToBoard(cursor.Point())
    a.tx = int(bx)
    a.ty = int(by)
  }
  if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    ex,ey := a.ent.Pos()
    if dist(ex, ey, a.tx, a.ty) <= a.Range && a.ent.HasLos(a.tx, a.ty) {
      a.ent.Stats.ApplyDamage(-a.Ap, 0, status.Unspecified)
      x := a.tx - (a.Diameter + 1) / 2
      y := a.ty - (a.Diameter + 1) / 2
      x2 := a.tx + a.Diameter / 2
      y2 := a.ty + a.Diameter / 2
      a.targets = nil
      for _,ent := range g.Ents {
        entx,enty := ent.Pos()
        if entx >= x && entx < x2 && enty >= y && enty < y2 {
          a.targets = append(a.targets, ent)
        }
      }
      return game.ConsumedAndBegin
    } else {
      return game.Consumed
    }
    // for _,target := range a.targets {
    //   if target == g.HoveredEnt() {
    //     if a.ent.Stats.ApCur() >= a.Ap && target.Stats.HpCur() > 0 {
    //       a.target = target
    //       a.ent.Stats.ApplyDamage(-a.Ap, 0, status.Unspecified)
    //       return game.ConsumedAndBegin
    //     }
    //     return game.Consumed
    //   }
    // }
    return game.Consumed
  }
  return game.NotConsumed
}
func (a *AoeAttack) RenderOnFloor() {
  ex,ey := a.ent.Pos()
  if dist(ex, ey, a.tx, a.ty) <= a.Range && a.ent.HasLos(a.tx, a.ty) {
    gl.Color4d(1.0, 0.2, 0.2, 0.8)
  } else {
    gl.Color4d(0.6, 0.6, 0.6, 0.8)
  }
  gl.Disable(gl.TEXTURE_2D)
  gl.Begin(gl.QUADS)
    gl.Vertex2i(a.tx - (a.Diameter + 1) / 2, a.ty - (a.Diameter + 1) / 2)
    gl.Vertex2i(a.tx - (a.Diameter + 1) / 2, a.ty + a.Diameter / 2)
    gl.Vertex2i(a.tx + a.Diameter / 2, a.ty + a.Diameter / 2)
    gl.Vertex2i(a.tx + a.Diameter / 2, a.ty - (a.Diameter + 1) / 2)
  gl.End()
}
func (a *AoeAttack) Cancel() {
  a.aoeAttackInst = aoeAttackInst{}
}
func (a *AoeAttack) Maintain(dt int64) game.MaintenanceStatus {
  if a.ent.Sprite.Sprite().State() != "ready" { return game.InProgress }
  for _,target := range a.targets {
    if target.Stats.HpCur() > 0 && target.Sprite.Sprite().State() != "ready" { return game.InProgress }
  }
  a.ent.TurnToFace(a.tx, a.ty)
  for _,target := range a.targets {
    target.TurnToFace(a.tx, a.ty)
  }
  a.ent.Sprite.Sprite().Command(a.Animation)
  for _,target := range a.targets {
    target.Sprite.Sprite().Command("defend")
    if game.DoAttack(a.ent, target, a.Strength, a.Kind) {
      for _,name := range a.Conditions {
        target.Stats.ApplyCondition(status.MakeCondition(name))
      }
      target.Stats.ApplyDamage(0, -a.Damage, a.Kind)
      if target.Stats.HpCur() <= 0 {
        target.Sprite.Sprite().Command("killed")
      } else {
        target.Sprite.Sprite().Command("damaged")
      }
    } else {
      target.Sprite.Sprite().Command("undamaged")
    }
  }
  return game.Complete
}
func (a *AoeAttack) Interrupt() bool {
  return true
}

