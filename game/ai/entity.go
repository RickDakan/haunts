package ai

import (
  "reflect"
  "github.com/runningwild/glop/ai"
  "github.com/runningwild/glop/util/algorithm"
  "github.com/runningwild/haunts/game/actions"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  "github.com/runningwild/haunts/house"
  "github.com/runningwild/polish"
)

func (a *Ai) addEntityContext(ent *game.Entity, context *polish.Context) {
  polish.AddFloat64MathContext(context)
  polish.AddBooleanContext(context)
  context.SetParseOrder(polish.Float, polish.String)
  a.addCommonContext(ent.Game())

  // This entity, the one currently taking its turn
  context.SetValue("me", ent)

  // All actions that the entity has are available using their names,
  // converted to lower case, and replacing spaces with underscores.
  // For example, "Kiss of Death" -> "kiss_of_death"

  // These functions are self-explanitory, they are all relative to the
  // current entity
  context.AddFunc("numVisibleEnemies",
      func() float64 {
        base.Log().Printf("Minion: numVisibleEnemies")
        side := game.SideHaunt
        if ent.Side() == game.SideHaunt {
          side = game.SideExplorers
        }
        return float64(numVisibleEntities(ent, side))
      })
  context.AddFunc("nearestEnemy",
      func() *game.Entity {
        base.Log().Printf("Minion: nearestEnemy")
        side := game.SideHaunt
        if ent.Side() == game.SideHaunt {
          side = game.SideExplorers
        }
        nearest := nearestEntity(ent, side)
        x1, y1 := ent.Pos()
        x2, y2 := nearest.Pos()
        base.Log().Printf("Nearest to (%d %d) is (%d %d)", x1, y1, x2, y2)
        return nearest
      })

  context.AddFunc("walkingDistBetween", walkingDistBetween)
  context.AddFunc("rangedDistBetween", rangedDistBetween)

  // Checks whether an entity is nil, this is important to check when using
  // function that returns an entity (like lastOffensiveTarget)
  context.AddFunc("stillExists", func(target *game.Entity) bool {
    base.Log().Printf("Minion: stillExists")
    return target != nil && target.Stats != nil && target.Stats.HpCur() > 0
  })

  // Returns the last entity that this ai attacked.  If the entity has died
  // this can return nil, so be sure to check that before using it.
  context.AddFunc("lastEntAttackedBy", func(ent *game.Entity) *game.Entity {
    base.Log().Printf("Minion: lastEntAttackedBy")
    return ent.Game().EntityById(ent.Info.LastEntThatIAttacked)
  })

  context.AddFunc("lastEntThatAttacked", func(ent *game.Entity) *game.Entity {
    base.Log().Printf("Minion: lastEntThatAttacked")
    return ent.Game().EntityById(ent.Info.LastEntThatAttackedMe)
  })

  context.AddFunc("advanceInRange", func(ents []*game.Entity, min_dist, max_dist, max_ap float64) {
    base.Log().Printf("Minion: advanceInRange")
    name := getActionName(ent, reflect.TypeOf(&actions.Move{}))
    move := getActionByName(ent, name).(*actions.Move)
    exec := move.AiMoveInRange(ent, ents, int(min_dist), int(max_dist), int(max_ap))
    if exec != nil {
      a.execs <- exec
      <-a.pause
    } else {
      a.graph.Term() <- ai.TermError
    }
  })

  context.AddFunc("group", func(e *game.Entity) []*game.Entity {
    base.Log().Printf("Minion: group")
    return []*game.Entity{e}
  })

  context.AddFunc("advanceAllTheWay", func(ents []*game.Entity) {
    base.Log().Printf("Minion: advanceAllTheWay")
    name := getActionName(ent, reflect.TypeOf(&actions.Move{}))
    move := getActionByName(ent, name).(*actions.Move)
    exec := move.AiMoveInRange(ent, ents, 1, 1, ent.Stats.ApCur())
    if exec != nil {
      a.execs <- exec
      <-a.pause
    } else {
      base.Log().Printf("Got a nil exec from move.AiMoveInRange")
      a.graph.Term() <- ai.TermError
    }
  })

  context.AddFunc("costToMoveInRange", func(ents []*game.Entity, min_dist, max_dist float64) float64 {
    base.Log().Printf("Minion: costToMoveInRange")
    name := getActionName(ent, reflect.TypeOf(&actions.Move{}))
    move := getActionByName(ent, name).(*actions.Move)
    cost := move.AiCostToMoveInRange(ent, ents, int(min_dist), int(max_dist))
    return float64(cost)
  })

  context.AddFunc("allIntruders", func() []*game.Entity {
    base.Log().Printf("Minion: allIntruders")
    var ents []*game.Entity
    for _, target := range ent.Game().Ents {
      if target.ExplorerEnt != nil {
        ents = append(ents, target)
      }
    }
    return ents
  })

  context.AddFunc("getBasicAttack", func() string {
    base.Log().Printf("Minion: getBasicAttack")
    return getActionName(ent, reflect.TypeOf(&actions.BasicAttack{}))
  })

  context.AddFunc("doBasicAttack", func(target *game.Entity, attack_name string) {
    base.Log().Printf("Minion: doBasicAttack")
    _attack := getActionByName(ent, attack_name)
    attack := _attack.(*actions.BasicAttack)
    exec := attack.AiAttackTarget(ent, target)
    if exec != nil {
      a.execs <- exec
      <-a.pause
    } else {
      a.graph.Term() <- ai.TermError
    }
  })

  context.AddFunc("basicAttackStat", func(action, stat string) float64 {
    base.Log().Printf("Minion: basicAttackStat")
    attack := getActionByName(ent, action).(*actions.BasicAttack)
    var val int
    switch stat {
      case "ap":
        val = attack.Ap
      case "damage":
        val = attack.Damage
      case "strength":
        val = attack.Strength
      case "range":
        val = attack.Range
      case "ammo":
        val = attack.Current_ammo
      default:
        base.Error().Printf("Requested basicAttackStat %s, which doesn't exist", stat)
    }
    return float64(val)
  })

  context.AddFunc("aoeAttackStat", func(action, stat string) float64 {
    base.Log().Printf("Minion: aoeAttackStat")
    attack := getActionByName(ent, action).(*actions.AoeAttack)
    var val int
    switch stat {
      case "ap":
        val = attack.Ap
      case "damage":
        val = attack.Damage
      case "strength":
        val = attack.Strength
      case "range":
        val = attack.Range
      case "ammo":
        val = attack.Current_ammo
      case "diameter":
        val = attack.Diameter
      default:
        base.Error().Printf("Requested aoeAttackStat %s, which doesn't exist", stat)
    }
    return float64(val)
  })

  context.AddFunc("master", func() *game.Entity {
    base.Log().Printf("Minion: master")
    for _, ent := range ent.Game().Ents {
      if ent.HauntEnt != nil && ent.HauntEnt.Level == game.LevelMaster {
        return ent
      }
    }
    return nil
  })

  context.AddFunc("nearestUnexploredRoom", func() *house.Room {
    g := ent.Game()
    graph := g.RoomGraph()
    current_room_num := ent.CurrentRoom()
    var unexplored []int
    for room_num, _ := range g.House.Floors[0].Rooms {
      if !ent.Info.RoomsExplored[room_num] {
        unexplored = append(unexplored, room_num)
      }
    }
    if len(unexplored) == 0 {
      return nil
    }
    cost, path := algorithm.Dijkstra(graph, []int{current_room_num}, unexplored)
    if cost == -1 {
      return nil
    }
    return g.House.Floors[0].Rooms[path[len(path) - 1]]
  })

  // Ends an entity's turn
  context.AddFunc("done", func() {
    base.Log().Printf("Minion: done")
    a.active_set <- false
  })
}

func numVisibleEntities(e *game.Entity, side game.Side) float64 {
  count := 0
  for _,ent := range e.Game().Ents {
    if ent == e { continue }
    if ent.Stats == nil || ent.Stats.HpCur() <= 0 { continue }
    if ent.Side() != side { continue }
    x,y := ent.Pos()
    if e.HasLos(x, y, 1, 1) {
      count++
    }
  }
  return float64(count)
}

