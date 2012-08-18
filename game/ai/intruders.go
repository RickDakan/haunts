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
  a.L.Register("SetEntityMasterInfo", setEntityMasterInfo(a))
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
      game.LuaPushEntity(L, ent)
      L.SetTable(-3)
    }
    base.Log().Printf("%d active intruders", count)
    return 1
  }
}

func setEntityMasterInfo(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    base.Log().Printf("Exec intruder")
    if !game.LuaCheckParamsOk(L, "SetEntityMasterInfo", game.LuaEntity, game.LuaString, game.LuaAnything) {
      return 0
    }
    ent := game.LuaToEntity(L, a.game, -3)
    if ent == nil {
      game.LuaDoError(L, "Tried to execIntruder on an invalid entity.")
      return 0
    }
    if ent.ExplorerEnt == nil {
      game.LuaDoError(L, "Tried to execIntruder on a non-intruder.")
      return 0
    }
    if ent.Ai_data == nil {
      ent.Ai_data = make(map[string]string)
    }
    if L.IsNil(-1) {
      delete(ent.Ai_data, L.ToString(-2))
    } else {
      ent.Ai_data[L.ToString(-2)] = L.ToString(-1)
    }
    return 0
  }
}

func execIntruderFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    base.Log().Printf("Exec intruder")
    if !game.LuaNumParamsOk(L, 1, "execIntruder") {
      return 0
    }
    ent := game.LuaToEntity(L, a.game, -1)
    if ent == nil {
      game.LuaDoError(L, "Tried to execIntruder on an invalid entity.")
      return 0
    }
    if ent.ExplorerEnt == nil {
      game.LuaDoError(L, "Tried to execIntruder on a non-intruder.")
      return 0
    }
    if !ent.Ai.Active() {
      game.LuaDoError(L, fmt.Sprintf("Tried to execIntruder '%s', who is not active.", ent.Name))
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
