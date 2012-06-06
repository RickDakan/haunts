package ai

import (
  "math/rand"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  "github.com/runningwild/haunts/house"
  "github.com/runningwild/polish"
)

func (a *Ai) addHigherContext(g *game.Game) {
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

func (a *Ai) addCommonContext(g *game.Game) {
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

  a.graph.Context.AddFunc("posOf", func(target *game.Entity) Pos {
    x, y := target.Pos()
    return Pos{x, y}
  })

  a.graph.Context.AddFunc("roomAt", func(pos Pos) *house.Room {
    return roomAt(pos, g)
  })

  a.graph.Context.AddFunc("doorAt", func(pos Pos) *house.Door {
    room := roomAt(pos, g)
    if room == nil {
      return nil
    }
    pos.X -= room.Size.Dx
    pos.Y -= room.Size.Dy
    for _, door := range room.Doors {
      switch door.Facing {
      case house.FarLeft:
        if pos.Y == room.Size.Dy - 1 && pos.X >= door.Pos - 1 && pos.X < door.Pos + door.Width + 1 {
          return door
        }
      case house.FarRight:
        if pos.X == room.Size.Dx - 1 && pos.Y >= door.Pos - 1 && pos.Y < door.Pos + door.Width + 1 {
          return door
        }
      case house.NearLeft:
        if pos.Y == 0 && pos.X >= door.Pos - 1 && pos.X < door.Pos + door.Width + 1 {
          return door
        }
      case house.NearRight:
        if pos.X == 0 && pos.Y >= door.Pos - 1 && pos.Y < door.Pos + door.Width + 1 {
          return door
        }
      }
    }
    return nil
  })

  a.graph.Context.AddFunc("roomExists", func(room *house.Room) bool {
    return room != nil
  })

  a.graph.Context.AddFunc("doorExists", func(door *house.Door) bool {
    return door != nil
  })

  a.graph.Context.AddFunc("doorOpened", func(door *house.Door) bool {
    return door.Opened
  })
}

func roomAt(pos Pos, g *game.Game) *house.Room {
  for _, room := range g.House.Floors[0].Rooms {
    if pos.X < room.X { continue }
    if pos.Y < room.Y { continue }
    if pos.X >= room.X + room.Size.Dx { continue }
    if pos.Y >= room.Y + room.Size.Dy { continue }
    return room
  }
  return nil
}


// This is the context used for the ai that controls the denizens' minions
func (a *Ai) addMinionsContext(g *game.Game) {
  polish.AddFloat64MathContext(a.graph.Context)
  polish.AddBooleanContext(a.graph.Context)
  a.graph.Context.SetParseOrder(polish.Float, polish.String)
  a.addCommonContext(g)
  a.addHigherContext(g)

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
