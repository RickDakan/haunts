package ai

import (
  "math/rand"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  "github.com/runningwild/polish"
)

func (a *Ai) addDenizensContext() {
  polish.AddFloat64MathContext(a.graph.Context)
  polish.AddBooleanContext(a.graph.Context)
  a.graph.Context.SetParseOrder(polish.Float, polish.String)
  a.addCommonContext()
  a.addHigherContext()

  // Returns the number of servitors that have not completed their turn
  a.graph.Context.AddFunc("numActiveServitors", func() float64 {
    count := 0.0
    for _, e := range a.game.Ents {
      if e.Side() != game.SideHaunt { continue }
      if e.HauntEnt.Level != game.LevelServitor { continue }
      if !e.Ai.Active() { continue }
      count++
    }
    return count
  })

  // Returns a random active servitor
  a.graph.Context.AddFunc("randomActiveServitor", func() *game.Entity {
    var ent *game.Entity
    count := 1.0
    for _, e := range a.game.Ents {
      if e.Side() != game.SideHaunt { continue }
      if e.HauntEnt.Level != game.LevelServitor { continue }
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

