package actions

import (
  "encoding/gob"
  "haunts/game/status"
  "haunts/game"
)

type ActionChargeAttack struct {
  Defname string
  *ActionChargeAttackDef
}
type ActionChargeAttackDef struct {
  Name     string
  Kind     status.Kind
  Ap       int
  Strength int
  Range    int
}

func init() {
  gob.Register(&ActionChargeAttack{})
}

func (a *ActionChargeAttack) Readyable() bool {
  return true
}
func (a *ActionChargeAttack) Cost() int {
  return a.Ap
}
func (a *ActionChargeAttack) Prep() bool {
  return true
}
func (a *ActionChargeAttack) HandleInput() bool {
  return true
}
func (a *ActionChargeAttack) HandleOutput() {
}
func (a *ActionChargeAttack) Cancel() {
}
func (a *ActionChargeAttack) Maintain(dt int64) game.MaintenanceStatus {
  // Do stuff
  return game.Complete
}
func (a *ActionChargeAttack) Interrupt() bool {
  return true
}

