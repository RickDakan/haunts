package ai2

import (
  "fmt"
  "github.com/runningwild/glop/util/algorithm"
  "github.com/runningwild/haunts/game/actions"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  "github.com/runningwild/haunts/house"
  lua "github.com/xenith-studios/golua"
  "sort"
)

func (a *Ai) addEntityContext() {
  a.L.Register("pos", posFunc(a))
  a.L.Register("me", meFunc(a))
  a.L.Register("allPathablePoints", allPathablePointsFunc(a))
  a.L.Register("rangedDistBetween", rangedDistBetweenFunc(a.ent))
  a.L.Register("nearestNEntities", nearestNEntitiesFunc(a.ent))
  a.L.Register("doBasicAttack", doBasicAttackFunc(a))
  a.L.Register("doMove", doMoveFunc(a))
}

type entityDist struct {
  dist int
  ent  *game.Entity
}
type entityDistSlice []entityDist
func (e entityDistSlice) Len() int {
  return len(e)
}
func (e entityDistSlice) Less(i, j int) bool {
  return e[i].dist < e[j].dist
}
func (e entityDistSlice) Swap(i, j int) {
  e[i], e[j] = e[j], e[i]
}

func getActionByName(e *game.Entity, name string) game.Action {
  for _,action := range e.Actions {
    if action.String() == name {
      return action
    }
  }
  return nil
}

// Assuming a table, t, on the top of the stack, returns t[x], t[y]
func getPointFromTable(L *lua.State) (int, int) {
  L.PushString("x")
  L.GetTable(-2)
  x := L.ToInteger(-1)
  L.Pop(1)
  L.PushString("y")
  L.GetTable(-2)
  y := L.ToInteger(-1)
  L.Pop(1)
  return x, y
}

// Makes a table with the keys x and y and leaves it on the top of the stack.
func putPointToTable(L *lua.State, x, y int) {
  L.NewTable()
  L.PushString("x")
  L.PushInteger(x)
  L.SetTable(-3)
  L.PushString("y")
  L.PushInteger(y)
  L.SetTable(-3)
}

// Used for doing los computation in the ai, so we don't have to allocate
// and deallocate lots of these.  Only one ai is ever running at a time so
// this should be ok.
var grid [][]bool
func init() {
  raw := make([]bool, house.LosTextureSizeSquared)
  grid = make([][]bool, house.LosTextureSize)
  for i := range grid {
    grid[i] = raw[i*house.LosTextureSize:(i+1)*house.LosTextureSize]
  }
}

// Input:
// 1 - Integer - Id of an entity.
// Output:
// 1 - table[x,y] - Position of the specified entity.
func posFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 1, "pos") {
      return 0
    }
    id := game.EntityId(L.ToInteger(-1))
    ent := a.ent.Game().EntityById(id)
    if ent == nil {
      luaDoError(L, fmt.Sprintf("Tried to get the position an entity with id=%d who doesn't exist.", id))
      return 0
    }
    x, y := ent.Pos()
    putPointToTable(L, x, y)
    return 1
  }
}


// Input: none
// Output:
// 1 - Integer - Id of the entity that this ai controls.
func meFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 0, "me") {
      return 0
    }
    L.PushInteger(int(a.ent.Id))
    return 1
  }
}

// Input:
// 1 - table[x,y] - Src, where the path starts from.
// 2 - table[x,y] - Dst, where the path ends if it goes to distance 0.
// 3 - Integer - Minimum distance from Dst that a path must end at.
// 4 - Integer - Maximum distance from Dst that a path must end at.
// Output:
// 1 - Table - An array of all of the points that a unit standing at Src can
//     reach that are at between the specified minimum and maximum distances
//     from Dst.  Remember that a valid pass cannot cross other entities,
//     furniture, walls, closed doors, or go out of what your los of the Src
//     position.
func allPathablePointsFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 4, "allPathablePoints") {
      return 0
    }
    if !L.IsTable(-4) || !L.IsTable(-3) || !L.IsNumber(-2) || !L.IsNumber(-1) {
      luaDoError(L, fmt.Sprintf("Unexpected parameters, expected allPathablePoints(table, table, int, int)"))
      return 0
    }
    min := L.ToInteger(-2)
    max := L.ToInteger(-1)
    L.Pop(2)
    x2, y2 := getPointFromTable(L)
    L.Pop(1)
    x1, y1 := getPointFromTable(L)
    L.Pop(1)

    a.ent.Game().DetermineLos(x2, y2, max, grid)
    var dst []int
    for x := x2 - max; x <= x2 + max; x++ {
      for y := y2 - max; y <= y2 + max; y++ {
        if x > x2 - min && x < x2 + min && y > y2 - min && y < y2 + min {
          continue
        }
        dst = append(dst, a.ent.Game().ToVertex(x, y))
      }
    }
    graph := a.ent.Game().Graph(nil)
    src := []int{a.ent.Game().ToVertex(x1, y1)}
    reachable := algorithm.ReachableDestinations(graph, src, dst)
    L.NewTable()
    for i, v := range reachable {
      _, x, y := a.ent.Game().FromVertex(v)
      L.PushInteger(i+1)
      putPointToTable(L, x, y)
      L.SetTable(-3)
    }
    return 1
  }
}

// Input:
// 1 - String - Name of the attack to use
// 2 - EntityId - Id of the target
// Output:
// 1 - Table - If the action was successful a table with the following keys
//     will be returned:
//     "Hit": Boolean indicated whether or not the attack hit its target
//     This table will be nil if the action was invalid for some reason.
func doBasicAttackFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    base.Log().Printf("Do basic attack")
    if !luaNumParamsOk(L, 2, "doBasicAttack") {
      return 0
    }
    me := a.ent
    name := L.ToString(-2)
    action := getActionByName(me, name)
    if action == nil {
      luaDoError(L, fmt.Sprintf("Entity '%s' (id=%d) has no action named '%s'.", me.Name, me.Id, name))
      return 0
    }
    id := game.EntityId(L.ToInteger(-1))
    target := me.Game().EntityById(id)
    if action == nil {
      luaDoError(L, fmt.Sprintf("Tried to target entity with id=%d who doesn't exist.", id))
      return 0
    }
    base.Log().Printf("Do basic attack - should execute at this point")
    attack := action.(*actions.BasicAttack)
    exec := attack.AiAttackTarget(me, target)
    if exec != nil {
      a.execs <- exec
      <-a.pause
      result := actions.GetBasicAttackResult(exec)
      if result == nil {
        L.PushNil()
      } else {
        L.NewTable()
        L.PushString("hit")
        L.PushBoolean(result.Hit)
        L.SetTable(-3)
      }
    } else {
      L.PushNil()
    }
    return 1
  }
}

// Input:
// 1 - table[table[x,y]] - Array of acceptable destinations
// 2 - Integer - Maximum ap to spend while doing this move.
// Output:
// 1 - table[x,y] - New position of this entity, or nil if the move failed.
func doMoveFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 2, "doMove") {
      return 0
    }
    if !L.IsTable(-2) || !L.IsNumber(-1) {
      luaDoError(L, fmt.Sprintf("Unexpected parameters, expected doMove(table, int)"))
      return 0
    }
    me := a.ent
    max_ap := L.ToInteger(-1)
    L.Pop(1)
    n := int(L.ObjLen(-1))
    dsts := make([]int, n)[0:0]
    for i := 1; i <= n; i++ {
      L.PushInteger(i)
      L.GetTable(-2)
      x, y := getPointFromTable(L)
      dsts = append(dsts, me.Game().ToVertex(x, y))
      L.Pop(1)
    }
    var move *actions.Move
    var ok bool
    for i := range me.Actions {
      move, ok = me.Actions[i].(*actions.Move)
      if ok {
        break
      }
    }
    if !ok {
      // TODO: what to do here?  This poor guy didn't have a move action :(
      L.PushNil()
      return 1
    }
    exec := move.AiMoveToPos(me, dsts, max_ap)
    if exec != nil {
      a.execs <- exec
      <-a.pause
      // TODO: Need to get a resolution
      x, y := me.Pos()
      putPointToTable(L, x, y)
      base.Log().Printf("Finished move")
    } else {
      base.Log().Printf("Didn't bother moving")
      L.PushNil()
    }
    return 1
  }
}


// Input:
// 1 - EntityId - Id of one entity in Los
// 2 - EntityId - Id another entity in Los
// Output:
// 1 - Integer - The ranged distance between the two entities.  If either of
//     the entities specified are not in los this function will return 10000.
func rangedDistBetweenFunc(me *game.Entity) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 2, "rangedDistBetween") {
      return 0
    }
    id1 := game.EntityId(L.ToInteger(-2))
    id2 := game.EntityId(L.ToInteger(-1))
    e1 := me.Game().EntityById(id1)
    e2 := me.Game().EntityById(id2)
    if e1 == nil || e2 == nil {
      L.PushInteger(10000)
    } else {
      L.PushInteger(rangedDistBetween(e1, e2))
    }
    return 1
  }
}

// Input:
// 1 - Integer - Maximum number of entities to return
// 2 - String  - One of "intruder" "denizen" "minion" "servitor" "master"
//     "non-minion" "non-servitor" "non-master"
//     The "non-*" parameters indicate denizens only (i.e. will *not* include
//     intruders) that are not of the type specified.
// Output:
// 1 - Table - Contains a mapping from index to entity Id, sorted in ascending
//     order of distance from the entity that called this function.  Only ents
//     that this unit has los to will be included in the output.
func nearestNEntitiesFunc(me *game.Entity) lua.GoFunction {
  valid_kinds := map[string]bool {
    "intruder": true,
    "denizen": true,
    "minion": true,
    "servitor": true,
    "master": true,
    "non-minion": true,
    "non-servitor": true,
    "non-master": true,
  }
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 2, "nearestNEntities") {
      return 0
    }
    g := me.Game()
    max := L.ToInteger(-2)
    kind := L.ToString(-1)
    if !valid_kinds[kind] {
      err_str := fmt.Sprintf("nearestNEntities expects kind in the set ['intruder' 'denizen' 'servitor' 'master' 'minion' 'non-servitor' 'non-master' 'non-minion'], got %s.", kind)
      base.Warn().Printf(err_str)
      L.PushString(err_str)
      L.Error()
      return 0
    }
    var eds entityDistSlice
    for _, ent := range g.Ents {
      switch kind {
      case "intruder":
        if ent.Side() != game.SideExplorers { continue }
      case "denizen":
        if ent.Side() != game.SideHaunt { continue }
      case "minion":
        if ent.HauntEnt == nil || ent.HauntEnt.Level != game.LevelMinion { continue }
      case "servitor":
        if ent.HauntEnt == nil || ent.HauntEnt.Level != game.LevelServitor { continue }
      case "master":
        if ent.HauntEnt == nil || ent.HauntEnt.Level != game.LevelMaster { continue }
      case "non-minion":
        if ent.HauntEnt == nil || ent.HauntEnt.Level == game.LevelMinion { continue }
      case "non-servitor":
        if ent.HauntEnt == nil || ent.HauntEnt.Level == game.LevelServitor { continue }
      case "non-master":
        if ent.HauntEnt == nil || ent.HauntEnt.Level == game.LevelMaster { continue }
      }
      x, y := ent.Pos()
      dx, dy := ent.Dims()
      if !me.HasLos(x, y, dx, dy) {
        continue
      }
      eds = append(eds, entityDist{rangedDistBetween(me, ent), ent})
    }
    sort.Sort(eds)
    if max > len(eds) {
      max = len(eds)
    }
    if max < 0 {
      max = 0
    }
    eds = eds[0:max]

    // eds contains the results, in order.  Now we make a lua table and
    // populate it with the entity ids of the results.
    L.NewTable()
    for i := range eds {
      L.PushInteger(i+1)
      L.PushInteger(int(eds[i].ent.Id))
      L.SetTable(-3)
    }
    return 1
  }
}

// func (a *Ai) addEntityContext(ent *game.Entity, context *polish.Context) {
//   polish.AddFloat64MathContext(context)
//   polish.AddBooleanContext(context)
//   context.SetParseOrder(polish.Float, polish.String)
//   a.addCommonContext(ent.Game())

//   // This entity, the one currently taking its turn
//   context.SetValue("me", ent)

//   // All actions that the entity has are available using their names,
//   // converted to lower case, and replacing spaces with underscores.
//   // For example, "Kiss of Death" -> "kiss_of_death"

//   // These functions are self-explanitory, they are all relative to the
//   // current entity
//   context.AddFunc("numVisibleEnemies",
//       func() float64 {
//         base.Log().Printf("Minion: numVisibleEnemies")
//         side := game.SideHaunt
//         if ent.Side() == game.SideHaunt {
//           side = game.SideExplorers
//         }
//         return float64(numVisibleEntities(ent, side))
//       })
//   context.AddFunc("nearestEnemy",
//       func() *game.Entity {
//         base.Log().Printf("Minion: nearestEnemy")
//         side := game.SideHaunt
//         if ent.Side() == game.SideHaunt {
//           side = game.SideExplorers
//         }
//         nearest := nearestEntity(ent, side)
//         x1, y1 := ent.Pos()
//         x2, y2 := nearest.Pos()
//         base.Log().Printf("Nearest to (%d %d) is (%d %d)", x1, y1, x2, y2)
//         return nearest
//       })

//   context.AddFunc("walkingDistBetween", walkingDistBetween)
//   context.AddFunc("rangedDistBetween", rangedDistBetween)

//   // Checks whether an entity is nil, this is important to check when using
//   // function that returns an entity (like lastOffensiveTarget)
//   context.AddFunc("stillExists", func(target *game.Entity) bool {
//     base.Log().Printf("Minion: stillExists")
//     return target != nil && target.Stats != nil && target.Stats.HpCur() > 0
//   })

//   // Returns the last entity that this ai attacked.  If the entity has died
//   // this can return nil, so be sure to check that before using it.
//   context.AddFunc("lastEntAttackedBy", func(ent *game.Entity) *game.Entity {
//     base.Log().Printf("Minion: lastEntAttackedBy")
//     return ent.Game().EntityById(ent.Info.LastEntThatIAttacked)
//   })

//   context.AddFunc("lastEntThatAttacked", func(ent *game.Entity) *game.Entity {
//     base.Log().Printf("Minion: lastEntThatAttacked")
//     return ent.Game().EntityById(ent.Info.LastEntThatAttackedMe)
//   })

//   context.AddFunc("advanceInRange", func(ents []*game.Entity, min_dist, max_dist, max_ap float64) {
//     base.Log().Printf("Minion: advanceInRange")
//     name := getActionName(ent, reflect.TypeOf(&actions.Move{}))
//     move := getActionByName(ent, name).(*actions.Move)
//     exec := move.AiMoveInRange(ent, ents, int(min_dist), int(max_dist), int(max_ap))
//     if exec != nil {
//       a.execs <- exec
//       <-a.pause
//     } else {
//       a.graph.Term() <- ai.TermError
//     }
//   })

//   context.AddFunc("group", func(e *game.Entity) []*game.Entity {
//     base.Log().Printf("Minion: group")
//     return []*game.Entity{e}
//   })

//   context.AddFunc("advanceAllTheWay", func(ents []*game.Entity) {
//     base.Log().Printf("Minion: advanceAllTheWay")
//     name := getActionName(ent, reflect.TypeOf(&actions.Move{}))
//     move := getActionByName(ent, name).(*actions.Move)
//     exec := move.AiMoveInRange(ent, ents, 1, 1, ent.Stats.ApCur())
//     if exec != nil {
//       a.execs <- exec
//       <-a.pause
//     } else {
//       base.Log().Printf("Got a nil exec from move.AiMoveInRange")
//       a.graph.Term() <- ai.TermError
//     }
//   })

//   context.AddFunc("costToMoveInRange", func(ents []*game.Entity, min_dist, max_dist float64) float64 {
//     base.Log().Printf("Minion: costToMoveInRange")
//     name := getActionName(ent, reflect.TypeOf(&actions.Move{}))
//     move := getActionByName(ent, name).(*actions.Move)
//     cost := move.AiCostToMoveInRange(ent, ents, int(min_dist), int(max_dist))
//     return float64(cost)
//   })

//   context.AddFunc("allIntruders", func() []*game.Entity {
//     base.Log().Printf("Minion: allIntruders")
//     var ents []*game.Entity
//     for _, target := range ent.Game().Ents {
//       if target.ExplorerEnt != nil {
//         ents = append(ents, target)
//       }
//     }
//     return ents
//   })

//   context.AddFunc("getBasicAttack", func() string {
//     base.Log().Printf("Minion: getBasicAttack")
//     return getActionName(ent, reflect.TypeOf(&actions.BasicAttack{}))
//   })

//   context.AddFunc("doBasicAttack", func(target *game.Entity, attack_name string) {
//     base.Log().Printf("Minion: doBasicAttack")
//     _attack := getActionByName(ent, attack_name)
//     attack := _attack.(*actions.BasicAttack)
//     exec := attack.AiAttackTarget(ent, target)
//     if exec != nil {
//       a.execs <- exec
//       <-a.pause
//     } else {
//       a.graph.Term() <- ai.TermError
//     }
//   })

//   context.AddFunc("basicAttackStat", func(action, stat string) float64 {
//     base.Log().Printf("Minion: basicAttackStat")
//     attack := getActionByName(ent, action).(*actions.BasicAttack)
//     var val int
//     switch stat {
//       case "ap":
//         val = attack.Ap
//       case "damage":
//         val = attack.Damage
//       case "strength":
//         val = attack.Strength
//       case "range":
//         val = attack.Range
//       case "ammo":
//         val = attack.Current_ammo
//       default:
//         base.Error().Printf("Requested basicAttackStat %s, which doesn't exist", stat)
//     }
//     return float64(val)
//   })

//   context.AddFunc("aoeAttackStat", func(action, stat string) float64 {
//     base.Log().Printf("Minion: aoeAttackStat")
//     attack := getActionByName(ent, action).(*actions.AoeAttack)
//     var val int
//     switch stat {
//       case "ap":
//         val = attack.Ap
//       case "damage":
//         val = attack.Damage
//       case "strength":
//         val = attack.Strength
//       case "range":
//         val = attack.Range
//       case "ammo":
//         val = attack.Current_ammo
//       case "diameter":
//         val = attack.Diameter
//       default:
//         base.Error().Printf("Requested aoeAttackStat %s, which doesn't exist", stat)
//     }
//     return float64(val)
//   })

//   context.AddFunc("master", func() *game.Entity {
//     base.Log().Printf("Minion: master")
//     for _, ent := range ent.Game().Ents {
//       if ent.HauntEnt != nil && ent.HauntEnt.Level == game.LevelMaster {
//         return ent
//       }
//     }
//     return nil
//   })

//   context.AddFunc("nearestUnexploredRoom", func() *house.Room {
//     g := ent.Game()
//     graph := g.RoomGraph()
//     current_room_num := ent.CurrentRoom()
//     var unexplored []int
//     for room_num, _ := range g.House.Floors[0].Rooms {
//       if !ent.Info.RoomsExplored[room_num] {
//         unexplored = append(unexplored, room_num)
//       }
//     }
//     if len(unexplored) == 0 {
//       return nil
//     }
//     cost, path := algorithm.Dijkstra(graph, []int{current_room_num}, unexplored)
//     if cost == -1 {
//       return nil
//     }
//     return g.House.Floors[0].Rooms[path[len(path) - 1]]
//   })

//   context.AddFunc("nextRoomTowards", func(target *house.Room) *house.Room {
//     g := ent.Game()
//     graph := g.RoomGraph()
//     current_room_num := ent.CurrentRoom()
//     if g.House.Floors[0].Rooms[current_room_num] == target {
//       return target
//     }
//     target_num := -1
//     for room_num, room := range g.House.Floors[0].Rooms {
//       if room == target {
//         target_num = room_num
//         break
//       }
//     }
//     if target_num == -1 {
//       return nil
//     }
//     cost, path := algorithm.Dijkstra(graph, []int{current_room_num}, []int{target_num})
//     if cost == -1 {
//       return nil
//     }
//     return g.House.Floors[0].Rooms[path[1]]
//   })

//   // Ends an entity's turn
//   context.AddFunc("done", func() {
//     base.Log().Printf("Minion: done")
//     a.active_set <- false
//   })
// }

// func numVisibleEntities(e *game.Entity, side game.Side) float64 {
//   count := 0
//   for _,ent := range e.Game().Ents {
//     if ent == e { continue }
//     if ent.Stats == nil || ent.Stats.HpCur() <= 0 { continue }
//     if ent.Side() != side { continue }
//     x,y := ent.Pos()
//     if e.HasLos(x, y, 1, 1) {
//       count++
//     }
//   }
//   return float64(count)
// }

