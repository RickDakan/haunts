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
  a.L.Register("pos", PosFunc(a))
  a.L.Register("me", MeFunc(a))
  a.L.Register("allPathablePoints", AllPathablePointsFunc(a))
  a.L.Register("rangedDistBetweenPositions", RangedDistBetweenPositionsFunc(a.ent))
  a.L.Register("rangedDistBetweenEntities", RangedDistBetweenEntitiesFunc(a.ent))
  a.L.Register("nearestNEntities", NearestNEntitiesFunc(a.ent))
  a.L.Register("getBasicAttackStats", GetBasicAttackStatsFunc(a))
  a.L.Register("getEntityStats", GetEntityStatsFunc(a))
  a.L.Register("getConditions", GetConditionsFunc(a))
  a.L.Register("getActions", GetActionsFunc(a))
  a.L.Register("entityInfo", EntityInfoFunc(a))
  a.L.Register("doBasicAttack", DoBasicAttackFunc(a))
  a.L.Register("doMove", DoMoveFunc(a))
  a.L.Register("exists", ExistsFunc(a))
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

// Returns the position of the specified entity, which must be in los.
//    Format:
//    p = pos(id)
//
//    Inputs:
//    id - integer - Entity id of the entity whose position this should return.
//
//    Outputs:
//    p - table[x,y] - The position of the specified entity, or nil if the
//                     entity was not in los.
func PosFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 1, "pos") {
      return 0
    }
    id := game.EntityId(L.ToInteger(-1))
    ent := a.ent.Game().EntityById(id)
    if ent == nil {
      L.PushNil()
      return 1
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


// Returns the entity id of the entity that is being controlled by this ai.
//    Format:
//    myid = me()
//
//    Outputs:
//    myid - integer
func MeFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 0, "me") {
      return 0
    }
    L.PushInteger(int(a.ent.Id))
    return 1
  }
}

// Returns an array of all points that can be reached by walking from a
// specific location that end in a certain general area.  Assumes that a 1x1
// unit is doing the walking.
//    Format:
//    points = allPathablePoints(src, dst, min, max)
//
//    Inputs:
//    src - table[x,y] - Where the path starts.
//    dst - table[x,y] - Another point near where the path should go.
//    min - integer    - Minimum distance from dst that the path should end at.
//    max - integer    - Maximum distance from dst that the path should end at.
//
//    Outputs:
//    points - array[table[x,y]]
func AllPathablePointsFunc(a *Ai) lua.GoFunction {
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

// Gets some basic persistent data about an entity.
//    Format:
//    info = entityInfo(id)
//
//    Inputs:
//    id - integer - Entity id of some entity
//
//    Outputs:
//    info - table - Table containing the following values:
//                   lastEntityIAttacked      (integer)
//                   lastEntityThatAttackedMe (integer)
func EntityInfoFunc(a *Ai) lua.GoFunction {
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

// Performs a basic attack against the specifed target.
//    Format:
//    res = doBasicAttack(attack, target)
//
//    Inputs:
//    attack - string  - Name of the attack to use.
//    target - integer - Entity id of the target of this attack.
//
//    Outputs:
//    res - table - Table containing the following values:
//                  hit (boolean) - true iff the attack hit its target.
//                  If the attack was invalid for some reason res will be nil.
func DoBasicAttackFunc(a *Ai) lua.GoFunction {
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

// Performs a move action to the closest one of any of the specifed inputs
// points.  The movement can be restricted to not spend more than a certain
// amount of ap.
//    Format:
//    p = doMove(dsts, max_ap)
//
//    Input:
//    dsts  - array[table[x,y]] - Array of all points that are acceptable
//                                destinations.
//    max_ap - integer - Maxmium ap to spend while doing this move, if the
//                       required ap exceeds this the entity will still move
//                       as far as possible towards a destination.
//
//    Output:
//    p - table[x,y] - New position of this entity, or nil if the move failed.
func DoMoveFunc(a *Ai) lua.GoFunction {
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


// Computes the ranged distance between two points.
//    Format:
//    dist = rangedDistBetweenPositions(p1, p2)
//
//    Input:
//    p1 - table[x,y]
//    p2 - table[x,y]
//
//    Output:
//    dist - integer - The ranged distance between the two positions.
func RangedDistBetweenPositionsFunc(me *game.Entity) lua.GoFunction {
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

// Computes the ranged distance between two entities.
//    Format:
//    dist = rangedDistBetweenEntities(e1, e2)
//
//    Input:
//    e1 - integer - An entity id.
//    e2 - integer - Another entity id.
//
//    Output:
//    dist - integer - The ranged distance between the two specified entities,
//                     this will not necessarily be the same as
//                     rangedDistBetweenPositions(pos(e1), pos(e2)) if at
//                     least one of the entities isn't 1x1.
func RangedDistBetweenEntitiesFunc(me *game.Entity) lua.GoFunction {
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

// Gets some stats about a basic attack.  If the specified action is not a
// basic attack this function will return nil.  It is an error to query an
// entity for an action that it does not have.
//    Format:
//    stats = getBasicAttackStats(id, name)
//
//    Input:
//    id   - integer - Entity id of the entity with the action to query.
//    name - string  - Name of the action to query.
//
//    Output:
//    stats - table - Table containing the following values:
//                    ap       (integer)
//                    damage   (integer)
//                    strength (integer)
//                    range    (integer)
//                    ammo     (integer)
func GetBasicAttackStatsFunc(a *Ai) lua.GoFunction {
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
          // It's not a basic attack, that's ok but we have to return nil.
          L.PushNil()
          return 1
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

// Gets some stats about an entity.
//    Format:
//    stats = getBasicAttackStats(id)
//
//    Input:
//    id   - integer - Entity id of the entity to query.
//
//    Output:
//    stats - table - Table containing the following values:
//                    corpus (integer)
//                    ego    (integer)
//                    hpCur  (integer)
//                    hpMax  (integer)
//                    apCur  (integer)
//                    apMax  (integer)
func GetEntityStatsFunc(a *Ai) lua.GoFunction {
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

// Returns a list of all conditions affecting an entity.
//    Format:
//    conditions = getConditions(id)
//
//    Input:
//    id - integer - Entity id of the entity whose conditions to query.
//
//    Output:
//    conditions - table - Contains a mapping from name to a true boolean for
//                         every condition currently affecting the specified
//                         entity.
func GetConditionsFunc(a *Ai) lua.GoFunction {
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

// Returns a list of all actions that an entity has.
//    Format:
//    actions = getActions(id)
//
//    Input:
//    id - integer - Entity id of the entity whose actions to query.
//
//    Output:
//    actions - table - Contains a mapping from name to a true boolean for
//                      every action that the specified entity has.
func GetActionsFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaNumParamsOk(L, 1, "getActions") {
      return 0
    }
    if !L.IsNumber(-1) {
      luaDoError(L, fmt.Sprintf("Unexpected parameters, expected getActions(int)"))
      return 0
    }
    id := game.EntityId(L.ToInteger(-1))
    ent := a.ent.Game().EntityById(id)
    if ent == nil {
      luaDoError(L, fmt.Sprintf("Tried to reference entity with id=%d who doesn't exist.", id))
      return 0
    }
    L.NewTable()
    for _, action := range ent.Actions {
      L.PushString(action.String())
      L.PushBoolean(true)
      L.SetTable(-3)
    }
    return 1
  }
}

// Queries whether or not an entity still exists.  An entity existing implies
// that it currently alive.
//    Format:
//    e = exists(id)
//
//    Input:
//    id - integer - Entity id of the entity whose existence we are querying.
//
//    Output:
//    e - boolean - True if the entity exists and has positive hp.
func ExistsFunc(a *Ai) lua.GoFunction {
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
    L.PushBoolean(ent != nil && ent.Stats != nil && ent.Stats.HpCur() > 0)
    return 1
  }
}

// Returns an array of all entities of a specified type that are in this
// entity's los.  The entities in the array will be sorted in ascending order
// of distance from this entity.
//    Format
//    ents = nearestNEntites(max, kind)
//
//    Input:
//    max  - integer - Maximum number of entities to return
//    kind - string  - One of "intruder" "denizen" "minion" "servitor"
//                     "master" "non-minion" "non-servitor" "non-master" and
//                     "all".  The "non-*" parameters indicate denizens only
//                     (i.e. will *not* include intruders) that are not of the
//                     type specified.
//    Output:
//    ents - array[integer] - Array of entity ids.
func NearestNEntitiesFunc(me *game.Entity) lua.GoFunction {
  valid_kinds := map[string]bool {
    "intruder": true,
    "denizen": true,
    "minion": true,
    "servitor": true,
    "master": true,
    "non-minion": true,
    "non-servitor": true,
    "non-master": true,
    "all" : true,
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
      if ent.Stats == nil { continue }
      switch kind {
      case "all":
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
