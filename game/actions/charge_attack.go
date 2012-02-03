package actions

import (
  "encoding/gob"
  "path/filepath"
  "haunts/base"
  "haunts/game/status"
  "haunts/game"
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

func init() {
  game.RegisterActionMakers(registerCharges)
  gob.Register(&ChargeAttack{})
}

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
}

func (a *ChargeAttack) Readyable() bool {
  return true
}
func (a *ChargeAttack) Cost() int {
  return a.Ap
}
func (a *ChargeAttack) Prep(*game.Entity) bool {
  return true
}
func (a *ChargeAttack) HandleInput() bool {
  return true
}
func (a *ChargeAttack) HandleOutput() {
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

