package ai

import (
  "fmt"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  lua "github.com/xenith-studios/golua"
)

func (a *Ai) addMinionsContext() {
  a.L.Register("activeMinions", activeMinionsFunc(a))
  a.L.Register("execMinion", execMinionFunc(a))
}

// Input:
//   None
// Output:
// 1 - Table - Contains a mapping from index to entity Id of all active
//     minions.
func activeMinionsFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 0, "activeMinions") {
      return 0
    }
    L.NewTable()
    count := 0
    for _, ent := range a.game.Ents {
      if ent.HauntEnt == nil || ent.HauntEnt.Level != game.LevelMinion { continue }
      if !ent.Ai.Active() { continue }
      count++
      L.PushInteger(count)
      L.PushInteger(int(ent.Id))
      L.SetTable(-3)
    }
    base.Log().Printf("activeMinions: %d", count)
    return 1
  }
}

func execMinionFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    base.Log().Printf("Exec minion")
    if !luaNumParamsOk(L, 1, "execMinion") {
      return 0
    }
    id := game.EntityId(L.ToInteger(0))
    ent := a.game.EntityById(id)
    if ent == nil {
      luaDoError(L, fmt.Sprintf("Tried to execMinion entity with Id=%d, which doesn't exist.", id))
      return 0
    }
    if ent.HauntEnt == nil || ent.HauntEnt.Level != game.LevelMinion {
      luaDoError(L, fmt.Sprintf("Tried to execMinion entity with Id=%d, which is not a minion.", id))
      return 0
    }
    if !ent.Ai.Active() {
      luaDoError(L, fmt.Sprintf("Tried to execMinion entity with Id=%d, which is not active.", id))
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
