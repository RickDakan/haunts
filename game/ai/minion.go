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
    if !game.LuaNumParamsOk(L, 0, "activeMinions") {
      return 0
    }
    L.NewTable()
    count := 0
    for _, ent := range a.game.Ents {
      if ent.HauntEnt == nil || ent.HauntEnt.Level != game.LevelMinion {
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
    base.Log().Printf("activeMinions: %d", count)
    return 1
  }
}

func execMinionFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    base.Log().Printf("Exec minion")
    if !game.LuaNumParamsOk(L, 1, "execMinion") {
      return 0
    }
    ent := game.LuaToEntity(L, a.game, -1)
    if ent == nil {
      game.LuaDoError(L, "Tried to execMinion entity which doesn't exist.")
      return 0
    }
    if ent.HauntEnt == nil || ent.HauntEnt.Level != game.LevelMinion {
      game.LuaDoError(L, fmt.Sprintf("Tried to execMinion entity with Id=%d, which is not a minion.", ent.Id))
      return 0
    }
    if !ent.Ai.Active() {
      game.LuaDoError(L, fmt.Sprintf("Tried to execMinion entity with Id=%d, which is not active.", ent.Id))
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
