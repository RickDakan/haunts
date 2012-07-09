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
    if !game.LuaNumParamsOk(L, 0, "activeNonMinions") {
      return 0
    }
    L.NewTable()
    count := 0
    for _, ent := range a.game.Ents {
      if ent.HauntEnt == nil || ent.HauntEnt.Level == game.LevelMinion {
        continue
      }
      if !ent.Ai.Active() {
        continue
      }
      count++
      L.PushInteger(count)
      game.LuaPushEntity(L, ent)
      L.SetTable(-3)
    }
    return 1
  }
}

func execNonMinionFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    base.Log().Printf("Exec non-minion")
    if !game.LuaNumParamsOk(L, 1, "execNonMinion") {
      return 0
    }
    ent := game.LuaToEntity(L, a.game, -1)
    if ent == nil {
      game.LuaDoError(L, "Tried to execNonMinion an invalid entity.")
      return 0
    }
    if ent.HauntEnt == nil || ent.HauntEnt.Level == game.LevelMinion {
      game.LuaDoError(L, fmt.Sprintf("Tried to execNonMinion entity with Id=%d, which is a minion.", ent.Id))
      return 0
    }
    if !ent.Ai.Active() {
      game.LuaDoError(L, fmt.Sprintf("Tried to execNonMinion entity with Id=%d, which is not active.", ent.Id))
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
