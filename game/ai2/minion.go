package ai2

import (
  "fmt"
  // "github.com/runningwild/haunts/game/actions"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  // "github.com/runningwild/haunts/house"
  lua "github.com/xenith-studios/golua"
)

func (a *Ai) addMinionsContext() {
  a.L.Register("activeMinions", activeMinionsFunc(a))
  a.L.Register("execMinions", execMinionFunc(a))
}

// Input:
//   None
// Output:
// 1 - Table - Contains a mapping from index to entity Id of all active
//     minions.
func activeMinionsFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    n := L.GetTop()
    if n != 0 {
      errstr := fmt.Sprintf("activeMinions expects exactly 0 parameters, got %d.", n)
      base.Warn().Printf(errstr)
      L.PushString(errstr)
      L.Error()
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
    return 1
  }
}

func execMinionFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    n := L.GetTop()
    if n != 1 {
      errstr := fmt.Sprintf("execMinion expects exactly 1 parameters, got %d.", n)
      base.Warn().Printf(errstr)
      L.PushString(errstr)
      L.Error()
      return 0
    }
    id := game.EntityId(L.ToInteger(0))
    ent := a.game.EntityById(id)
    if ent == nil {
      base.Warn().Printf("Tried to execMinion entity with Id=%d, which doesn't exist.", id)
      return 0
    }
    if ent.HauntEnt == nil || ent.HauntEnt.Level != game.LevelMinion {
      base.Warn().Printf("Tried to execMinion entity with Id=%d, which is not a minion.", id)
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
