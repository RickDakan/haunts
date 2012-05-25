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
type aoeExec struct {
  game.BasicActionExec
  Pos int
}
func init() {
  gob.Register(aoeExec{})
}

func (a *AoeAttack) AP() int {
  return a.Ap
}
func (a *AoeAttack) Pos() (int, int) {
  return a.tx, a.ty
}
func (a *AoeAttack) Dims() (int, int) {
  return a.Diameter, a.Diameter
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
func (a *AoeAttack) HandleInput(group gui.EventGroup, g *game.Game) (bool, game.ActionExec) {
  cursor := group.Events[0].Key.Cursor()
  if cursor != nil && cursor.Name() == "Mouse" {
    bx,by := g.GetViewer().WindowToBoard(cursor.Point())
    a.tx = int(bx)
    a.ty = int(by)
  }
  if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    ex,ey := a.ent.Pos()
    if dist(ex, ey, a.tx, a.ty) <= a.Range && a.ent.HasLos(a.tx, a.ty, 1, 1) {
      var exec aoeExec
      exec.SetBasicData(a.ent, a)
      exec.Pos = a.ent.Game().ToVertex(a.tx, a.ty)
      return true, exec
    } else {
      return true, nil
    }
    return true, nil
  }
  return false, nil
}
func (a *AoeAttack) RenderOnFloor() {
  if a.ent == nil {
    return
  }
  ex,ey := a.ent.Pos()
  if dist(ex, ey, a.tx, a.ty) <= a.Range && a.ent.HasLos(a.tx, a.ty, 1, 1) {
    gl.Color4ub(255, 255, 255, 200)
  } else {
    gl.Color4ub(255, 64, 64, 200)
  }
  base.EnableShader("box")
  base.SetUniformF("box", "dx", float32(a.Diameter))
  base.SetUniformF("box", "dy", float32(a.Diameter))
  base.SetUniformI("box", "temp_invalid", 0)
  x := a.tx - (a.Diameter + 1) / 2
  y := a.ty - (a.Diameter + 1) / 2
  (&texture.Object{}).Data().Render(float64(x), float64(y), float64(a.Diameter), float64(a.Diameter))
  base.EnableShader("")
}
func (a *AoeAttack) Cancel() {
  a.aoeAttackTempData = aoeAttackTempData{}
}
func (a *AoeAttack) Maintain(dt int64, g *game.Game, ae game.ActionExec) game.MaintenanceStatus {
  if ae != nil {
    exec := ae.(aoeExec)
    _, tx, ty := g.FromVertex(exec.Pos)
    x := tx - (a.Diameter + 1) / 2
    y := ty - (a.Diameter + 1) / 2
    x2 := tx + a.Diameter / 2
    y2 := ty + a.Diameter / 2
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
    a.targets = targets
    if a.Current_ammo > 0 {
      a.Current_ammo--
    }
    a.ent = g.EntityById(ae.EntityId())
    a.ent.Stats.ApplyDamage(-a.Ap, 0, status.Unspecified)

    // Track this information for the ais - the attacking ent will only
    // remember one ent that it hit, but that's ok
    for _, target := range a.targets {
      if target.Side() != a.ent.Side() {
        target.Info.LastEntThatAttackedMe = a.ent.Id
        a.ent.Info.LastEntThatIAttacked = target.Id
        break
      }
    }
  }
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

