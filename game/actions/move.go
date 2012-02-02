package actions

import (
  "haunts/base"
  "haunts/game"
  "encoding/gob"
  "path/filepath"
)

func registerMoves() map[string]func() game.Action {
  move_actions := make(map[string]*MoveDef)
  base.RemoveRegistry("actions-move_actions")
  base.RegisterRegistry("actions-move_actions", move_actions)
  base.RegisterAllObjectsInDir("actions-move_actions", filepath.Join(base.GetDataDir(), "actions", "movement"), ".json", "json")
  makers := make(map[string]func() game.Action)
  for name := range move_actions {
    cname := name
    makers[cname] = func() game.Action {
      a := Move{ Defname: cname }
      base.GetObject("actions-move_actions", &a)
      return &a
    }
  }
  return makers
}

func init() {
  game.RegisterActionMakers(registerMoves)
  gob.Register(Move{})
}

type Move struct {
  Defname string
  *MoveDef
}
type MoveDef struct {
  Name     string
}

func (a *Move) Readyable() bool {
  return true
}
func (a *Move) Cost() int {
  return 3
}
func (a *Move) Prep(*game.Entity) bool {
  return true
}
func (a *Move) HandleInput() bool {
  return true
}
func (a *Move) HandleOutput() {
}
func (a *Move) Cancel() {
}
func (a *Move) Maintain(dt int64) game.MaintenanceStatus {
  // Do stuff
  return game.Complete
}
func (a *Move) Interrupt() bool {
  return true
}

