package ai

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
  a.L.Register("rangedDistBetweenPositions", rangedDistBetweenPositionsFunc(a.ent))
  a.L.Register("rangedDistBetweenEntities", rangedDistBetweenEntitiesFunc(a.ent))
  a.L.Register("nearestNEntities", nearestNEntitiesFunc(a.ent))
  a.L.Register("getBasicAttackStats", getBasicAttackStatsFunc(a))
  a.L.Register("getEntityStats", getEntityStatsFunc(a))
  a.L.Register("getConditions", getConditionsFunc(a))
  a.L.Register("entityInfo", entityInfoFunc(a))
  a.L.Register("doBasicAttack", doBasicAttackFunc(a))
  a.L.Register("doMove", doMoveFunc(a))
  a.L.Register("exists", existsFunc(a))
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
// 1 - table[x,y] - Position of the specified entity if the entity is in los,
//     otherwise it will return nil.
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
    if a.ent.HasLos(x, y, 1, 1) {
      putPointToTable(L, x, y)
    } else {
      L.PushNil()
    }
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
// 1 - Integer - Id of the entity we are querying
// Output:
// 1 - Table - The following keys are populated in the return value:
//     lastEntityIAttacked: Id of the last entity that this one attacked
//     lastEntityThatAttackedMe: Id of the last entity to attack this one
//
//     Note - Just because an Id is present in the table does not mean that
//     its corresponding entity still exists, you still need to check that it
//     exists with the exists() function.
func entityInfoFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 1, "entityInfo") {
      return 0
    }
    if !L.IsNumber(-1) {
      luaDoError(L, fmt.Sprintf("Unexpected parameters, expected entityInfo(int)"))
      return 0
    }
    id := game.EntityId(L.ToInteger(-1))
    ent := a.ent.Game().EntityById(id)
    if ent == nil {
      luaDoError(L, fmt.Sprintf("Tried to reference entity with id=%d who doesn't exist.", id))
      return 0
    }
    L.NewTable()
    L.PushString("lastEntityIAttacked")
    L.PushInteger(int(ent.Info.LastEntThatIAttacked))
    L.SetTable(-3)
    L.PushString("lastEntityThatAttackedMe")
    L.PushInteger(int(ent.Info.LastEntThatAttackedMe))
    L.SetTable(-3)
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
// 1 - table[x,y] - One position
// 2 - table[x,y] - Another position
// Output:
// 1 - Integer - The ranged distance between the two positions.
func rangedDistBetweenPositionsFunc(me *game.Entity) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 2, "rangedDistBetweenPositions") {
      return 0
    }
    x1, y1 := getPointFromTable(L)
    L.Pop(1)
    x2, y2 := getPointFromTable(L)
    dx := x2 - x1
    if dx < 0 {
      dx = -dx
    }
    dy := y2 - y1
    if dy < 0 {
      dy = -dy
    }
    if dx > dy {
      L.PushInteger(dx)
    } else {
      L.PushInteger(dy)
    }
    return 1
  }
}

// Input:
// 1 - EntityId - Id of one entity in Los
// 2 - EntityId - Id another entity in Los
// Output:
// 1 - Integer - The ranged distance between the two entities.  If either of
//     the entities specified are not in los this function will return nil.
func rangedDistBetweenEntitiesFunc(me *game.Entity) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 2, "rangedDistBetweenEntities") {
      return 0
    }
    id1 := game.EntityId(L.ToInteger(-2))
    id2 := game.EntityId(L.ToInteger(-1))
    e1 := me.Game().EntityById(id1)
    if e1 == nil {
      L.PushNil()
      return 1
    }
    x, y := e1.Pos()
    dx, dy := e1.Dims()
    if !me.HasLos(x, y, dx, dy) {
      L.PushNil()
      return 1
    }
    e2 := me.Game().EntityById(id2)
    if e2 == nil {
      L.PushNil()
      return 1
    }

    x, y = e2.Pos()
    dx, dy = e2.Dims()
    if !me.HasLos(x, y, dx, dy) {
      L.PushNil()
      return 1
    }

    L.PushInteger(rangedDistBetween(e1, e2))
    return 1
  }
}

// Input:
// 1 - Integer - Entity id of the entity whose action we want to query
// 2 - String  - Name of the basic attack
// Output:
// 1 - Table - Contains a mapping from stat to value of that stat, includes
//     the following values: ap, damage, strength, range, ammo
func getBasicAttackStatsFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 2, "getBasicAttackStats") {
      return 0
    }
    if !L.IsNumber(-2) || !L.IsString(-1) {
      luaDoError(L, fmt.Sprintf("Unexpected parameters, expected getBasicAttackStats(int, string)"))
      return 0
    }
    id := game.EntityId(L.ToInteger(-2))
    ent := a.ent.Game().EntityById(id)
    if ent == nil {
      luaDoError(L, fmt.Sprintf("Tried to reference entity with id=%d who doesn't exist.", id))
      return 0
    }
    name := L.ToString(-1)
    for _, action := range ent.Actions {
      if action.String() == name {
        attack, ok := action.(*actions.BasicAttack)
        if !ok {
          luaDoError(L, fmt.Sprintf("%s is not a BasicAttack", name))
          return 0
        }
        L.NewTable()
        L.PushString("ap")
        L.PushInteger(attack.Ap)
        L.SetTable(-3)
        L.PushString("damage")
        L.PushInteger(attack.Damage)
        L.SetTable(-3)
        L.PushString("strength")
        L.PushInteger(attack.Strength)
        L.SetTable(-3)
        L.PushString("range")
        L.PushInteger(attack.Range)
        L.SetTable(-3)
        L.PushString("ammo")
        L.PushInteger(attack.Ammo)
        L.SetTable(-3)
        return 1
      }
    }
    luaDoError(L, fmt.Sprintf("Entity with id=%d has no action named %s", id, name))
    return 0
  }
}

// Input:
// 1 - Integer - Entity id of the entity whose action we want to query
// Output:
// 1 - Table - Contains a mapping from stat to value of that stat, includes
//     the following values: corpus, ego, apMax, apCur, hpMax, hpCur
func getEntityStatsFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 1, "getEntityStats") {
      return 0
    }
    if !L.IsNumber(-1) {
      luaDoError(L, fmt.Sprintf("Unexpected parameters, expected getEntityStats(int)"))
      return 0
    }
    id := game.EntityId(L.ToInteger(-1))
    ent := a.ent.Game().EntityById(id)
    if ent == nil {
      luaDoError(L, fmt.Sprintf("Tried to reference entity with id=%d who doesn't exist.", id))
      return 0
    }
    if ent.Stats == nil {
      luaDoError(L, fmt.Sprintf("Tried to query stats for entity with id=%d who doesn't have stats.", id))
      return 0
    }

    L.NewTable()
    L.PushString("corpus")
    L.PushInteger(ent.Stats.Corpus())
    L.SetTable(-3)
    L.PushString("ego")
    L.PushInteger(ent.Stats.Ego())
    L.SetTable(-3)
    L.PushString("hpCur")
    L.PushInteger(ent.Stats.HpCur())
    L.SetTable(-3)
    L.PushString("hpMax")
    L.PushInteger(ent.Stats.HpMax())
    L.SetTable(-3)
    L.PushString("apCur")
    L.PushInteger(ent.Stats.ApCur())
    L.SetTable(-3)
    L.PushString("apMax")
    L.PushInteger(ent.Stats.ApMax())
    L.SetTable(-3)

    return 1
  }
}

// Input:
// 1 - Integer - Entity id of the entity whose conditions we want to know.
// Output:
// 1 - Table - Contains a mapping from condition name to a boolean value that
//     is set to true.
func getConditionsFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 1, "getConditions") {
      return 0
    }
    if !L.IsNumber(-1) {
      luaDoError(L, fmt.Sprintf("Unexpected parameters, expected getConditions(int)"))
      return 0
    }
    id := game.EntityId(L.ToInteger(-1))
    ent := a.ent.Game().EntityById(id)
    if ent == nil {
      luaDoError(L, fmt.Sprintf("Tried to reference entity with id=%d who doesn't exist.", id))
      return 0
    }
    L.NewTable()
    for _, condition := range ent.Stats.ConditionNames() {
      L.PushString(condition)
      L.PushBoolean(true)
      L.SetTable(-3)
    }
    return 1
  }
}

// Input:
// 1 - Integer - Entity id of the entity whose existence we are querying.
// Output:
// 1 - Boolean - True if the entity exists.
func existsFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 1, "exists") {
      return 0
    }
    if !L.IsNumber(-1) {
      luaDoError(L, fmt.Sprintf("Unexpected parameters, expected exists(int)"))
      return 0
    }
    id := game.EntityId(L.ToInteger(-1))
    ent := a.ent.Game().EntityById(id)
    L.PushBoolean(ent != nil)
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
