package ai

import (
  "math/rand"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  "github.com/runningwild/polish"
)

func (a *Ai) addHigherContext() {
  // Begins or continues executing an entity's ai and executes one action from
  // it.
  a.graph.Context.AddFunc("exec", func(ent *game.Entity) {
    base.Log().Printf("Execute %p", ent)
    exec := <-ent.Ai.ActionExecs()
    base.Log().Printf("Got an action: %v", exec)
    if exec != nil {
      base.Log().Printf("Sending that action")
      a.execs <- exec
      base.Log().Printf("Sent.")
    }
    <-a.pause
  })

  // Indicates that the entities are all done executing for this turn.
  a.graph.Context.AddFunc("done", func() {
    base.Log().Printf("master done")
    <-a.pause
  })
}

func (a *Ai) addCommonContext() {
  // rolls dice, for example "roll 3 6" is a roll of 3d6
  a.graph.Context.AddFunc("roll", roll)

  a.graph.Context.AddFunc("entityStat", func(target *game.Entity, stat string) float64 {
    var val float64
    switch stat {
      case "corpus":
        val = float64(target.Stats.Corpus())
      case "ego":
        val = float64(target.Stats.Ego())
      case "hpMax":
        val = float64(target.Stats.HpMax())
      case "apMax":
        val = float64(target.Stats.ApMax())
      case "hpCur":
        val = float64(target.Stats.HpCur())
      case "apCur":
        val = float64(target.Stats.ApCur())
      default:
        base.Error().Printf("Requested entityStat %s, which doesn't exist", stat)
    }
    return val
  })

  a.graph.Context.AddFunc("hasCondition", func(target *game.Entity, name string) bool {
    for _, con := range target.Stats.ConditionNames() {
      if lowerAndUnderscore(con) == name {
        return true
      }
    }
    return false
  })

  a.graph.Context.AddFunc("hasAction", func(target *game.Entity, name string) bool {
    for _, action := range target.Actions {
      if lowerAndUnderscore(action.String()) == name {
        return true
      }
    }
    return false
  })
}

// This is the context used for the ai that controls the denizens' minions
func (a *Ai) addMinionsContext() {
  polish.AddFloat64MathContext(a.graph.Context)
  polish.AddBooleanContext(a.graph.Context)
  a.graph.Context.SetParseOrder(polish.Float, polish.String)
  a.addCommonContext()
  a.addHigherContext()

  // Returns the number of minions that have not completed their turn
  a.graph.Context.AddFunc("numActiveMinions", func() float64 {
    count := 0.0
    for _, e := range a.game.Ents {
      if e.Side() != game.SideHaunt { continue }
      if e.HauntEnt.Level != game.LevelMinion { continue }
      if !e.Ai.Active() { continue }
      count++
    }
    base.Log().Printf("Num active minions: %f", count)
    return count
  })

  // Returns a random active minion
  a.graph.Context.AddFunc("randomActiveMinion", func() *game.Entity {
    base.Log().Printf("randomActiveMinion")
    var ent *game.Entity
    count := 1.0
    for _, e := range a.game.Ents {
      if e.Side() != game.SideHaunt { continue }
      if e.HauntEnt.Level != game.LevelMinion { continue }
      if !e.Ai.Active() { continue }
      if rand.Float64() < 1.0 / count {
        ent = e
      }
      count++
    }
    base.Log().Printf("Selected %s (%p)", ent.Name, ent)
    return ent
  })
}
