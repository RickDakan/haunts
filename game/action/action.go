package action

import (
  "fmt"
  "haunts/base"
  "path/filepath"
)

type MaintenanceStatus int
const (
  InProgress         MaintenanceStatus = iota
  CheckForInterrupts
  Complete
)

var action_map map[string]func() Action

func MakeAction(name string) Action {
  f,ok := action_map[name]
  if !ok {
    panic(fmt.Sprintf("Unable to find an Action named '%s'", name))
  }
  return f()
}

func LoadAllActionsInDir(dir string) {
  action_map = make(map[string]func() Action)

  basic_attacks := make(map[string]*ActionBasicAttackDef)
  base.RemoveRegistry("actions-basic_attacks")
  base.RegisterRegistry("actions-basic_attacks", basic_attacks)
  s,_ := filepath.Abs(dir)
  fmt.Printf("abs: %s\n", s)
  base.RegisterAllObjectsInDir("actions-basic_attacks", filepath.Join(dir, "basic_attacks"), ".json", "json")
  for name := range basic_attacks {
    if _,ok := action_map[name]; ok {
      panic(fmt.Sprintf("Found two different actions with the same name: '%s'", name))
    }
    cname := name
    action_map[cname] = func() Action {
      a := ActionBasicAttack{ Defname: cname }
      base.GetObject("actions-basic_attacks", &a)
      return &a
    }
  }

  charge_attacks := make(map[string]*ActionChargeAttackDef)
  base.RemoveRegistry("actions-charge_attacks")
  base.RegisterRegistry("actions-charge_attacks", charge_attacks)
  base.RegisterAllObjectsInDir("actions-charge_attacks", filepath.Join(dir, "charge_attacks"), ".json", "json")
  for name := range charge_attacks {
    if _,ok := action_map[name]; ok {
      panic(fmt.Sprintf("Found two different actions with the same name: '%s'", name))
    }
    cname := name
    action_map[cname] = func() Action {
      a := ActionChargeAttack{ Defname: cname }
      base.GetObject("actions-charge_attacks", &a)
      return &a
    }
  }
}

type Action interface {
  // Returns true iff this action can be used as an interrupt
  Readyable() bool

  // Cost, in Ap, that must be paid to use this action.
  Cost() int

  // Called when the user first selects this action
  // Returns true if the action can be performed
  // Returns false if the action cannot be performed
  Prep() bool

  // Got to have some way for the user to interact with the action.  Returns
  // true if the action has been comitted.  If this action is not being
  // readied then it will take effect immediately.
  HandleInput() bool

  // Got to have some way for the user to see what is going on
  HandleOutput()

  // Called if the user cancels the action - done this way so that all actions
  // can be cancelled in the same way instead of each action deciding how to
  // cancel itself.
  Cancel()

  // Actually executes the action.  Returns a value after every call
  // indicating whether the action is done, still in progress, or can be
  // interrupted.
  Maintain(dt int64) MaintenanceStatus

  // This will be called if the action has been readied at this is a logical
  // point for an interrupt to happen.  Should return true if the action
  // should take place.
  Interrupt() bool
}
