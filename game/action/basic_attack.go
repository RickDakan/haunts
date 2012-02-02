package action

import (
  "encoding/gob"
  "haunts/game/status"
)

// Basic Attacks are single target and instant, they are also readyable
type ActionBasicAttack struct {
  Defname string
  *ActionBasicAttackDef
}
type ActionBasicAttackDef struct {
  Name     string
  Kind     status.Kind
  Ap       int
  Strength int
  Range    int
}

func init() {
  gob.Register(ActionBasicAttack{})
}

func (a *ActionBasicAttack) Readyable() bool {
  return true
}
func (a *ActionBasicAttack) Cost() int {
  return a.Ap
}
func (a *ActionBasicAttack) Prep() bool {
  return true
}
func (a *ActionBasicAttack) HandleInput() bool {
  return true
}
func (a *ActionBasicAttack) HandleOutput() {
}
func (a *ActionBasicAttack) Cancel() {
}
func (a *ActionBasicAttack) Maintain(dt int64) MaintenanceStatus {
  // Do stuff
  return Complete
}
func (a *ActionBasicAttack) Interrupt() bool {
  return true
}

