package ai

import (
  "math/rand"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  "github.com/runningwild/polish"
)

func (a *Ai) addIntrudersContext(g *game.Game) {
  polish.AddFloat64MathContext(a.graph.Context)
  polish.AddBooleanContext(a.graph.Context)
  a.graph.Context.SetParseOrder(polish.Float, polish.String)
  a.addCommonContext(g)
  a.addHigherContext(g)

  // Returns the number of intruders that have not completed their turn
  a.graph.Context.AddFunc("numActiveIntruders", func() float64 {
    count := 0.0
    for _, e := range a.game.Ents {
      if e.Side() != game.SideExplorers { continue }
      if !e.Ai.Active() { continue }
      count++
    }
    return count
  })

  // Returns a random active servitor
  a.graph.Context.AddFunc("randomActiveIntruder", func() *game.Entity {
    var ent *game.Entity
    count := 1.0
    for _, e := range a.game.Ents {
      if e.Side() != game.SideExplorers { continue }
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

