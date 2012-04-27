package actions

import (
  // "encoding/gob"
  "path/filepath"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/house"
  "github.com/runningwild/haunts/game"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/haunts/game/status"
)

func registerCharges() map[string]func() game.Action {
  charge_actions := make(map[string]*ChargeAttackDef)
  base.RemoveRegistry("actions-charge_actions")
  base.RegisterRegistry("actions-charge_actions", charge_actions)
  base.RegisterAllObjectsInDir("actions-charge_actions", filepath.Join(base.GetDataDir(), "actions", "charge_attacks"), ".json", "json")
  makers := make(map[string]func() game.Action)
  for name := range charge_actions {
    cname := name
    makers[cname] = func() game.Action {
      a := ChargeAttack{ Defname: cname }
      base.GetObject("actions-charge_actions", &a)
      return &a
    }
  }
  return makers
}

// func init() {
//   game.RegisterActionMakers(registerCharges)
//   gob.Register(&ChargeAttack{})
// }

type ChargeAttack struct {
  Defname string
  *ChargeAttackDef
}
type ChargeAttackDef struct {
  Name     string
  Kind     status.Kind
  Ap       int
  Strength int
  Range    int
  Texture  texture.Object
}

func (a *ChargeAttack) AP() int {
  return a.Ap
}
func (a *ChargeAttack) String() string {
  return a.Name
}
func (a *ChargeAttack) Icon() *texture.Object {
  return &a.Texture
}
func (a *ChargeAttack) Readyable() bool {
  return true
}
func (a *ChargeAttack) Preppable(ent *game.Entity, g *game.Game) bool {
  return true
}
func (a *ChargeAttack) Prep(ent *game.Entity, g *game.Game) bool {
  return true
}
func (a *ChargeAttack) HandleInput(gui.EventGroup, *game.Game) game.InputStatus {
  return game.NotConsumed
}
func (a *ChargeAttack) RenderOnFloor(room *house.Room) {
}
func (a *ChargeAttack) Cancel() {
}
func (a *ChargeAttack) Maintain(dt int64) game.MaintenanceStatus {
  // Do stuff
  return game.Complete
}
func (a *ChargeAttack) Interrupt() bool {
  return true
}

