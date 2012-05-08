package actions

import (
  "encoding/gob"
  "path/filepath"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/util/algorithm"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  "github.com/runningwild/haunts/game/status"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/opengl/gl"
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
      if a.Ammo > 0 {
        a.Current_ammo = a.Ammo
      } else {
        a.Current_ammo = -1
      }
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
  aoeAttackTempData

  Current_ammo int
}
type AoeAttackDef struct {
  Name       string
  Kind       status.Kind
  Ap         int
  Ammo       int  // 0 = infinity
  Strength   int
  Range      int
  Diameter   int
  Damage     int
  Animation  string
  Conditions []string
  Texture    texture.Object
}
type aoeAttackTempData struct {
  ent *game.Entity

  // position of the target cell - in the case of an even diameter this will be
  // the lower-left of the center
  tx,ty int

  // All entities in the blast radius - could include the acting entity
  targets []*game.Entity
}
func (a *AoeAttack) AP() int {
  return a.Ap
}
func (a *AoeAttack) Pos() (int, int) {
  return 0, 0
}
func (a *AoeAttack) Dims() (int, int) {
  return 0, 0
}
func (a *AoeAttack) String() string {
  return a.Name
}
func (a *AoeAttack) Icon() *texture.Object {
  return &a.Texture
}
func (a *AoeAttack) Readyable() bool {
  return true
}
func (a *AoeAttack) Preppable(ent *game.Entity, g *game.Game) bool {
  return a.Current_ammo != 0 && ent.Stats.ApCur() >= a.Ap
}
func (a *AoeAttack) Prep(ent *game.Entity, g *game.Game) bool {
  if !a.Preppable(ent, g) {
    return false
  }
  a.ent = ent
  bx,by := g.GetViewer().WindowToBoard(gin.In().GetCursor("Mouse").Point())
  a.tx = int(bx)
  a.ty = int(by)
  return true
}
func (a *AoeAttack) attack(targets []*game.Entity) {
  a.targets = targets
  if a.Current_ammo > 0 {
    a.Current_ammo--
  }
  a.ent.Stats.ApplyDamage(-a.Ap, 0, status.Unspecified)
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
    if dist(ex, ey, a.tx, a.ty) <= a.Range && a.ent.HasLos(a.tx, a.ty, 1, 1) {
      x := a.tx - (a.Diameter + 1) / 2
      y := a.ty - (a.Diameter + 1) / 2
      x2 := a.tx + a.Diameter / 2
      y2 := a.ty + a.Diameter / 2
      var targets []*game.Entity
      for _,ent := range g.Ents {
        entx,enty := ent.Pos()
        if entx >= x && entx < x2 && enty >= y && enty < y2 {
          targets = append(targets, ent)
        }
      }
      algorithm.Choose2(&targets, func(e *game.Entity) bool {
        return e.Stats != nil
      })
      a.attack(targets)
      return game.ConsumedAndBegin
    } else {
      return game.Consumed
    }
    return game.Consumed
  }
  return game.NotConsumed
}
func (a *AoeAttack) RenderOnFloor() {
  ex,ey := a.ent.Pos()
  if dist(ex, ey, a.tx, a.ty) <= a.Range && a.ent.HasLos(a.tx, a.ty, 1, 1) {
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
  a.aoeAttackTempData = aoeAttackTempData{}
}
func (a *AoeAttack) Maintain(dt int64) game.MaintenanceStatus {
  if a.ent.Sprite().State() != "ready" { return game.InProgress }
  for _,target := range a.targets {
    if target.Stats.HpCur() > 0 && target.Sprite().State() != "ready" { return game.InProgress }
  }
  a.ent.TurnToFace(a.tx, a.ty)
  for _,target := range a.targets {
    target.TurnToFace(a.tx, a.ty)
  }
  a.ent.Sprite().Command(a.Animation)
  for _,target := range a.targets {
    if game.DoAttack(a.ent, target, a.Strength, a.Kind) {
      for _,name := range a.Conditions {
        target.Stats.ApplyCondition(status.MakeCondition(name))
      }
      target.Stats.ApplyDamage(0, -a.Damage, a.Kind)
      if target.Stats.HpCur() <= 0 {
        target.Sprite().CommandN([]string{"defend", "killed"})
      } else {
        target.Sprite().CommandN([]string{"defend", "damaged"})
      }
    } else {
      target.Sprite().CommandN([]string{"defend", "undamaged"})
    }
  }
  return game.Complete
}
func (a *AoeAttack) Interrupt() bool {
  return true
}

