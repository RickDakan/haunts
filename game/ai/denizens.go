package ai

import (
  "fmt"
  // "github.com/runningwild/haunts/game/actions"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  // "github.com/runningwild/haunts/house"
  lua "github.com/xenith-studios/golua"
)

func (a *Ai) addDenizensContext() {
  a.L.Register("activeNonMinions", activeNonMinionsFunc(a))
  a.L.Register("execNonMinion", execNonMinionFunc(a))
}

// Input:
//   None
// Output:
// 1 - Table - Contains a mapping from index to entity Id of all active
//     non-minions.
func activeNonMinionsFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    base.Log().Printf("Acing non minions")
    if !luaNumParamsOk(L, 0, "activeNonMinions") {
      return 0
    }
    L.NewTable()
    count := 0
    for _, ent := range a.game.Ents {
      if ent.HauntEnt == nil || ent.HauntEnt.Level == game.LevelMinion { continue }
      base.Log().Printf("Servitor: %s", ent.Name)
      if !ent.Ai.Active() { continue }
      base.Log().Printf("Is active")
      count++
      L.PushInteger(count)
      L.PushInteger(int(ent.Id))
      L.SetTable(-3)
    }
    return 1
  }
}

func execNonMinionFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    base.Log().Printf("Exec non-minion")
    if !luaNumParamsOk(L, 1, "execNonMinion") {
      return 0
    }
    id := game.EntityId(L.ToInteger(0))
    ent := a.game.EntityById(id)
    if ent == nil {
      luaDoError(L, fmt.Sprintf("Tried to execNonMinion entity with Id=%d, which doesn't exist.", id))
      return 0
    }
    if ent.HauntEnt == nil || ent.HauntEnt.Level == game.LevelMinion {
      luaDoError(L, fmt.Sprintf("Tried to execNonMinion entity with Id=%d, which is a minion.", id))
      return 0
    }
    if !ent.Ai.Active() {
      luaDoError(L, fmt.Sprintf("Tried to execNonMinion entity with Id=%d, which is not active.", id))
      return 0
    }
    exec := <-ent.Ai.ActionExecs()
    if exec != nil {
      a.execs <- exec
    }
    <-a.pause
    return 0
  }
}
