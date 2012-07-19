package ai

import (
  "fmt"
  // "github.com/runningwild/haunts/game/actions"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  // "github.com/runningwild/haunts/house"
  lua "github.com/xenith-studios/golua"
)

func (a *Ai) addIntrudersContext() {
  a.L.Register("activeIntruders", activeIntrudersFunc(a))
  a.L.Register("execIntruder", execIntruderFunc(a))
}

// Input:
//   None
// Output:
// 1 - Table - Contains a mapping from index to entity Id of all active
//     intruders.
func activeIntrudersFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !game.LuaNumParamsOk(L, 0, "activeIntruders") {
      return 0
    }
    L.NewTable()
    count := 0
    for _, ent := range a.game.Ents {
      if ent.ExplorerEnt == nil {
        continue
      }
      if !ent.Ai.Active() {
        continue
      }
      count++
      L.PushInteger(count)
      L.PushInteger(int(ent.Id))
      L.SetTable(-3)
    }
    base.Log().Printf("%d active intruders", count)
    return 1
  }
}

func execIntruderFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    base.Log().Printf("Exec intruder")
    if !game.LuaNumParamsOk(L, 1, "execIntruder") {
      return 0
    }
    id := game.EntityId(L.ToInteger(0))
    ent := a.game.EntityById(id)
    if ent == nil {
      game.LuaDoError(L, fmt.Sprintf("Tried to execIntruder entity with Id=%d, which doesn't exist.", id))
      return 0
    }
    if ent.ExplorerEnt == nil {
      game.LuaDoError(L, fmt.Sprintf("Tried to execIntruder entity with Id=%d, which is not an intruder.", id))
      return 0
    }
    if !ent.Ai.Active() {
      game.LuaDoError(L, fmt.Sprintf("Tried to execIntruder entity with Id=%d, which is not active.", id))
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
