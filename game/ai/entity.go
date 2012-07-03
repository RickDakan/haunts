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
  a.loadUtils("entity")
  game.LuaPushEntity(L, a.ent)
  a.L.SetGlobal("Me")

  a.L.Register("pos", PosFunc(a))
  a.L.Register("allPathablePoints", AllPathablePointsFunc(a))
  a.L.Register("rangedDistBetweenPositions", RangedDistBetweenPositionsFunc(a))
  a.L.Register("rangedDistBetweenEntities", RangedDistBetweenEntitiesFunc(a))
  a.L.Register("nearestNEntities", NearestNEntitiesFunc(a.ent))

  // An entity table should have all of this.
  a.L.Register("getEntityStats", GetEntityStatsFunc(a))
  a.L.Register("getConditions", GetConditionsFunc(a))
  a.L.Register("getActions", GetActionsFunc(a))
  a.L.Register("entityInfo", EntityInfoFunc(a))
  a.L.Register("getBasicAttackStats", GetBasicAttackStatsFunc(a))
  a.L.Register("getAoeAttackStats", GetAoeAttackStatsFunc(a))

  a.L.Register("doBasicAttack", DoBasicAttackFunc(a))
  a.L.Register("doAoeAttack", DoAoeAttackFunc(a))
  a.L.Register("bestAoeAttackPos", BestAoeAttackPosFunc(a))
  a.L.Register("doMove", DoMoveFunc(a))
  a.L.Register("exists", ExistsFunc(a))

  a.L.Register("nearbyUnexploredRoom", NearbyUnexploredRoomFunc(a))
  a.L.Register("roomPath", RoomPathFunc(a))
  a.L.Register("roomContaining", RoomContainingFunc(a))
  a.L.Register("allDoorsBetween", AllDoorsBetween(a))
  a.L.Register("allDoorsOn", AllDoorsOn(a))
  a.L.Register("doorPositions", DoorPositionsFunc(a))
  a.L.Register("doorIsOpen", DoorIsOpenFunc(a))
  a.L.Register("doDoorToggle", DoDoorToggleFunc(a))
  a.L.Register("roomPositions", RoomPositionsFunc(a))
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
  for _, action := range e.Actions {
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
    grid[i] = raw[i*house.LosTextureSize : (i+1)*house.LosTextureSize]
  }
}

// Returns the position of the specified entity.
//    Format:
//    p = pos(id)
//
//    Inputs:
//    id - integer - Entity id of the entity whose position this should return.
//
//    Outputs:
//    p - table[x,y] - The position of the specified entity.
func PosFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "pos", luaInteger) {
      return 0
    }
    id := game.EntityId(L.ToInteger(-1))
    ent := a.ent.Game().EntityById(id)
    if ent == nil {
      L.PushNil()
      return 1
    }
    x, y := ent.Pos()
    putPointToTable(L, x, y)
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
    if !luaCheckParamsOk(L, "me") {
      return 0
    }

    game.LuaPushEntity(L, a.ent)
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
    if !luaCheckParamsOk(L, "allPathablePoints", luaTable, luaTable, luaInteger, luaInteger) {
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
    for x := x2 - max; x <= x2+max; x++ {
      for y := y2 - max; y <= y2+max; y++ {
        if x > x2-min && x < x2+min && y > y2-min && y < y2+min {
          continue
        }
        dst = append(dst, a.ent.Game().ToVertex(x, y))
      }
    }
    graph := a.ent.Game().Graph(a.ent.Side(), nil)
    src := []int{a.ent.Game().ToVertex(x1, y1)}
    reachable := algorithm.ReachableDestinations(graph, src, dst)
    L.NewTable()
    base.Log().Printf("%d reachable from (%d, %d) -> (%d, %d)", len(reachable), x1, y1, x2, y2)
    for i, v := range reachable {
      _, x, y := a.ent.Game().FromVertex(v)
      L.PushInteger(i + 1)
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
    if !luaCheckParamsOk(L, "entityInfo", luaInteger) {
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
    if !luaCheckParamsOk(L, "doBasicAttack", luaString, luaInteger) {
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
    attack, ok := action.(*actions.BasicAttack)
    if !ok {
      luaDoError(L, fmt.Sprintf("Action '%s' is not a basic attack.", name))
      return 0
    }
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

// Performs an aoe attack against centered at the specified position.
//    Format:
//    res = doAoeAttack(attack, pos)
//
//    Inputs:
//    attack - string     - Name of the attack to use.
//    pos    - table[x,y] - Position to center the aoe around.
//
//    Outputs:
//    res - boolean - true if the action performed, nil otherwise.
func DoAoeAttackFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "doAoeAttack", luaString, luaTable) {
      return 0
    }
    me := a.ent
    name := L.ToString(-2)
    action := getActionByName(me, name)
    if action == nil {
      luaDoError(L, fmt.Sprintf("Entity '%s' (id=%d) has no action named '%s'.", me.Name, me.Id, name))
      return 0
    }
    attack, ok := action.(*actions.AoeAttack)
    if !ok {
      luaDoError(L, fmt.Sprintf("Action '%s' is not an aoe attack.", name))
      return 0
    }
    tx, ty := getPointFromTable(L)
    exec := attack.AiAttackPosition(me, tx, ty)
    if exec != nil {
      a.execs <- exec
      <-a.pause
      L.PushBoolean(true)
    } else {
      L.PushNil()
    }
    return 1
  }
}

// Performs an aoe attack against centered at the specified position.
//    Format:
//    target = bestAoeAttackPos(attack, extra_dist, spec)
//
//    Inputs:
//    attack     - string  - Name of the attack to use.
//    extra_dist - integer - Available distance to move before attacking.
//    spec       - string  - One of the following values:
//                           "allies ok", "minions ok", "enemies only"
//
//    Outputs:
//    pos - table[x,y] - Position to place aoe for maximum results.
func BestAoeAttackPosFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "bestAoeAttackPos", luaString, luaInteger, luaString) {
      return 0
    }
    me := a.ent
    name := L.ToString(-3)
    action := getActionByName(me, name)
    if action == nil {
      luaDoError(L, fmt.Sprintf("Entity '%s' (id=%d) has no action named '%s'.", me.Name, me.Id, name))
      return 0
    }
    attack, ok := action.(*actions.AoeAttack)
    if !ok {
      luaDoError(L, fmt.Sprintf("Action '%s' is not an aoe attack.", name))
      return 0
    }
    var spec actions.AiAoeTarget
    switch L.ToString(-1) {
    case "allies ok":
      spec = actions.AiAoeHitAlliesOk
    case "minions ok":
      spec = actions.AiAoeHitMinionsOk
    case "enemies only":
      spec = actions.AiAoeHitNoAllies
    default:
      luaDoError(L, fmt.Sprintf("'%s' is not a valid value of spec for bestAoeAttackPos().", L.ToString(-1)))
      return 0
    }
    x, y := attack.AiBestTarget(me, L.ToInteger(-2), spec)
    putPointToTable(L, x, y)
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
    if !luaCheckParamsOk(L, "doMove", luaTable, luaInteger) {
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
func RangedDistBetweenPositionsFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "rangedDistBetweenPositions", luaTable, luaTable) {
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
func RangedDistBetweenEntitiesFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "rangedDistBetweenEntities", luaInteger, luaInteger) {
      return 0
    }
    id1 := game.EntityId(L.ToInteger(-2))
    id2 := game.EntityId(L.ToInteger(-1))
    e1 := a.ent.Game().EntityById(id1)
    if e1 == nil {
      L.PushNil()
      return 1
    }
    x, y := e1.Pos()
    dx, dy := e1.Dims()
    if !a.ent.HasLos(x, y, dx, dy) {
      L.PushNil()
      return 1
    }
    e2 := a.ent.Game().EntityById(id2)
    if e2 == nil {
      L.PushNil()
      return 1
    }

    x, y = e2.Pos()
    dx, dy = e2.Dims()
    if !a.ent.HasLos(x, y, dx, dy) {
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
    if !luaCheckParamsOk(L, "getBasicAttackStats", luaInteger, luaString) {
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
        if attack.Current_ammo == -1 {
          L.PushInteger(1000)
        } else {
          L.PushInteger(attack.Current_ammo)
        }
        L.SetTable(-3)
        return 1
      }
    }
    luaDoError(L, fmt.Sprintf("Entity with id=%d has no action named %s", id, name))
    return 0
  }
}

// Gets some stats about an aoe attack.  If the specified action is not an
// aoe attack this function will return nil.  It is an error to query an
// entity for an action that it does not have.
//    Format:
//    stats = getAoeAttackStats(id, name)
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
//                    diameter (integer)
//                    ammo     (integer)
func GetAoeAttackStatsFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "getAoeAttackStats", luaInteger, luaString) {
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
        attack, ok := action.(*actions.AoeAttack)
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
        L.PushString("diameter")
        L.PushInteger(attack.Diameter)
        L.SetTable(-3)
        L.PushString("ammo")
        if attack.Current_ammo == -1 {
          L.PushInteger(1000)
        } else {
          L.PushInteger(attack.Current_ammo)
        }
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
//    stats = getEntityStats(id)
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
    if !luaCheckParamsOk(L, "getEntityStats", luaInteger) {
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
    if !luaCheckParamsOk(L, "getConditions", luaInteger) {
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
    if !luaCheckParamsOk(L, "getActions", luaInteger) {
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
    if !luaCheckParamsOk(L, "exists", luaInteger) {
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
//
//    Output:
//    ents - array[integer] - Array of entity ids.
func NearestNEntitiesFunc(me *game.Entity) lua.GoFunction {
  valid_kinds := map[string]bool{
    "intruder":     true,
    "denizen":      true,
    "minion":       true,
    "servitor":     true,
    "master":       true,
    "non-minion":   true,
    "non-servitor": true,
    "non-master":   true,
    "all":          true,
  }
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "nearestNEntities", luaInteger, luaString) {
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
      if ent.Stats == nil {
        continue
      }
      if ent.Stats.HpCur() <= 0 {
        continue
      }
      switch kind {
      case "all":
      case "intruder":
        if ent.Side() != game.SideExplorers {
          continue
        }
      case "denizen":
        if ent.Side() != game.SideHaunt {
          continue
        }
      case "minion":
        if ent.HauntEnt == nil || ent.HauntEnt.Level != game.LevelMinion {
          continue
        }
      case "servitor":
        if ent.HauntEnt == nil || ent.HauntEnt.Level != game.LevelServitor {
          continue
        }
      case "master":
        if ent.HauntEnt == nil || ent.HauntEnt.Level != game.LevelMaster {
          continue
        }
      case "non-minion":
        if ent.HauntEnt == nil || ent.HauntEnt.Level == game.LevelMinion {
          continue
        }
      case "non-servitor":
        if ent.HauntEnt == nil || ent.HauntEnt.Level == game.LevelServitor {
          continue
        }
      case "non-master":
        if ent.HauntEnt == nil || ent.HauntEnt.Level == game.LevelMaster {
          continue
        }
      }
      x, y := ent.Pos()
      dx, dy := ent.Dims()
      if !me.HasLos(x, y, dx, dy) {
        continue
      }
      eds = append(eds, entityDist{rangedDistBetween(me, ent), ent})
    }
    // TODO: ONLY GUYS THAT EXIST
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
      L.PushInteger(i + 1)
      L.PushInteger(int(eds[i].ent.Id))
      L.SetTable(-3)
    }
    return 1
  }
}

func pushRoom(L *lua.State, floor, room int) {
  L.NewTable()
  L.PushString("floor")
  L.PushInteger(floor)
  L.SetTable(-3)
  L.PushString("room")
  L.PushInteger(room)
  L.SetTable(-3)
}

func pushDoor(L *lua.State, floor, room, door int) {
  L.NewTable()
  L.PushString("floor")
  L.PushInteger(floor)
  L.SetTable(-3)
  L.PushString("room")
  L.PushInteger(room)
  L.SetTable(-3)
  L.PushString("door")
  L.PushInteger(door)
  L.SetTable(-3)
}

func getFloorRoomDoor(L *lua.State, index int) (floor, room, door int) {
  L.PushString("floor")
  L.GetTable(index - 1)
  floor = L.ToInteger(-1)
  L.Pop(1)
  L.PushString("room")
  L.GetTable(index - 1)
  room = L.ToInteger(-1)
  L.Pop(1)
  L.PushString("door")
  L.GetTable(index - 1)
  door = L.ToInteger(-1)
  L.Pop(1)
  return
}

func checkFloorRoom(h *house.HouseDef, floor, room int) bool {
  if floor < 0 || room < 0 {
    return false
  }
  if floor >= len(h.Floors) {
    return false
  }
  if room >= len(h.Floors[floor].Rooms) {
    return false
  }
  return true
}

func checkFloorRoomDoor(h *house.HouseDef, floor, room, door int) bool {
  if !checkFloorRoom(h, floor, room) {
    return false
  }
  if door < 0 || door >= len(h.Floors[floor].Rooms[room].Doors) {
    return false
  }
  return true
}

// Returns one room that this entity has not explored that can be reached by
// going through only explored rooms.  It will return one of the closest such
// rooms.
//    Format
//    r = nearbyUnexploredRoom()
//
//    Input:
//    none
//
//    Output:
//    r - room - An unexplored room, or nil if no such room exists.
func NearbyUnexploredRoomFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "nearbyUnexploredRoom") {
      return 0
    }

    me := a.ent
    g := me.Game()
    graph := g.RoomGraph()
    current_room_num := me.CurrentRoom()
    var unexplored []int
    for room_num, _ := range g.House.Floors[0].Rooms {
      if !me.Info.RoomsExplored[room_num] {
        unexplored = append(unexplored, room_num)
      }
    }
    if len(unexplored) == 0 {
      base.Error().Printf("NO UNEXPLORED ROOMS!")
      L.PushNil()
      return 1
    }
    cost, path := algorithm.Dijkstra(graph, []int{current_room_num}, unexplored)
    if cost == -1 {
      base.Error().Printf("NO PATHABLE UNEXPLORED ROOMS!")
      L.PushNil()
      return 1
    }

    pushRoom(L, 0, path[len(path)-1])
    return 1
  }
}

// Returns a list of rooms representing a path from src to dst.  The path will
// not include src, but will include dst.  This function will return nil if
// the path requires going through more than a single unexplored room, this
// means that you can use this to path to an unexplored room, but you cannot
// use it to path to a room further in the house than that.
// rooms.
//    Format
//    path = roomPath(src, dst)
//
//    Input:
//    src - Room to start the path from.
//    dst - Room to end the path at.
//
//    Output:
//    path - array - A list of rooms that connect src to dst, excluding src
//    but including dst.
func RoomPathFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "roomPath", luaTable, luaTable) {
      return 0
    }

    me := a.ent
    g := me.Game()
    graph := g.RoomGraph()
    f1, r1, _ := getFloorRoomDoor(L, -2)
    if !checkFloorRoom(g.House, f1, r1) {
      luaDoError(L, fmt.Sprintf("Referenced floor and room (%d, %d) which doesn't exist.", f1, r1))
      return 0
    }
    f2, r2, _ := getFloorRoomDoor(L, -1)
    if !checkFloorRoom(g.House, f2, r2) {
      luaDoError(L, fmt.Sprintf("Referenced floor and room (%d, %d) which doesn't exist.", f2, r2))
      return 0
    }

    cost, path := algorithm.Dijkstra(graph, []int{r1}, []int{r2})
    if cost == -1 {
      L.PushNil()
      return 1
    }
    num_unexplored := 0
    for _, v := range path {
      if !me.Info.RoomsExplored[v] {
        num_unexplored++
      }
    }
    if num_unexplored > 1 {
      L.PushNil()
      return 1
    }
    L.NewTable()
    for i, v := range path {
      if i == 0 {
        continue
      } // Skip this one because we're in it already
      L.PushInteger(i)
      pushRoom(L, 0, v)
      L.SetTable(-3)
    }
    return 1
  }
}

// Returns the room that the specified entity is currently in.  The specified
// entity must be in los of a unit on the acting entity's team, or be on the
// acting entity's team, otherwise this function returns nil.
//    Format
//    r = roomContaining(id)
//
//    Input:
//    id - An entity id.
//
//    Output:
//    r - room - The room the specified entity is in, or nil if it can't be
//    seen right now.
func RoomContainingFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "roomContaining", luaInteger) {
      return 0
    }

    id := game.EntityId(L.ToInteger(-1))
    ent := a.ent.Game().EntityById(id)
    if ent == nil || (ent.Side() != a.ent.Side() && !a.ent.Game().TeamLos(ent.Pos())) {
      L.PushNil()
    } else {
      pushRoom(L, 0, ent.CurrentRoom())
    }
    return 1
  }
}

// Returns a list of all doors between two rooms.
//    Format
//    doors = allDoorsBetween(r1, r2)
//
//    Input:
//    r1 - room - A room.
//    r2 - room - Another room.
//
//    Output:
//    doors - array[door] - List of all doors connecting r1 and r2.
func AllDoorsBetween(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "allDoorsBetween", luaTable, luaTable) {
      return 0
    }
    f1, r1, _ := getFloorRoomDoor(L, -1)
    if !checkFloorRoom(a.ent.Game().House, f1, r1) {
      luaDoError(L, fmt.Sprintf("Referenced floor and room (%d, %d) which doesn't exist.", f1, r1))
      return 0
    }
    f2, r2, _ := getFloorRoomDoor(L, -2)
    if !checkFloorRoom(a.ent.Game().House, f2, r2) {
      luaDoError(L, fmt.Sprintf("Referenced floor and room (%d, %d) which doesn't exist.", f2, r2))
      return 0
    }
    if f1 != f2 {
      // Rooms on different floors can theoretically be connected in the
      // future by a stairway, but right now that doesn't happen.
      L.NewTable()
      return 1
    }

    L.NewTable()
    count := 1
    room1 := a.ent.Game().House.Floors[f1].Rooms[r1]
    room2 := a.ent.Game().House.Floors[f2].Rooms[r2]
    base.Log().Printf("Room1: (%d, %d) dims (%d, %d)", room1.X, room1.Y, room1.Size.Dx, room1.Size.Dy)
    base.Log().Printf("Room2: (%d, %d) dims (%d, %d)", room2.X, room2.Y, room2.Size.Dx, room2.Size.Dy)
    for d_index, door1 := range room1.Doors {
      for _, door2 := range room2.Doors {
        _, d := a.ent.Game().House.Floors[f1].FindMatchingDoor(room1, door1)
        if d == door2 {
          L.PushInteger(count)
          count++
          pushDoor(L, f1, r1, d_index)
          L.SetTable(-3)
        }
      }
    }
    return 1
  }
}

// Returns a list of all doors attached to the specified room.
//    Format
//    room = allDoorsOn(r)
//
//    Input:
//    r - room - A room.
//
//    Output:
//    doors - array[door] - List of all doors attached to the specified room.
func AllDoorsOn(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "allDoorsOn", luaTable) {
      return 0
    }
    f, r, _ := getFloorRoomDoor(L, -1)
    if !checkFloorRoom(a.ent.Game().House, f, r) {
      luaDoError(L, fmt.Sprintf("Referenced floor and room (%d, %d) which doesn't exist.", f, r))
      return 0
    }

    L.NewTable()
    for i := range a.ent.Game().House.Floors[f].Rooms[r].Doors {
      L.PushInteger(i + 1)
      pushDoor(L, f, r, i)
      L.SetTable(-3)
    }
    return 1
  }
}

// Returns a list of all positions that the specified door can be opened and
// closed from.
//    Format
//    ps = doorPositions(d)
//
//    Input:
//    d - door - A door.
//
//    Output:
//    ps - array[table[x,y]] - List of all position this door can be opened
//    and closed from.
func DoorPositionsFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "doorPositions", luaTable) {
      return 0
    }
    f, r, d := getFloorRoomDoor(L, -1)
    if !checkFloorRoomDoor(a.ent.Game().House, f, r, d) {
      luaDoError(L, fmt.Sprintf("Referenced floor, room, and door (%d, %d, %d) which doesn't exist.", f, r, d))
      return 0
    }
    room := a.ent.Game().House.Floors[f].Rooms[r]
    door := room.Doors[d]

    var x, y, dx, dy int
    switch door.Facing {
    case house.FarLeft:
      x = door.Pos
      y = room.Size.Dy
      dx = 1
    case house.FarRight:
      x = room.Size.Dy
      y = door.Pos
      dy = 1
    case house.NearLeft:
      x = -1
      y = door.Pos
      dy = 1
    case house.NearRight:
      x = door.Pos
      y = -1
      dx = 1
    default:
      luaDoError(L, fmt.Sprintf("Found a door with a bad facing."))
    }
    L.NewTable()
    count := 1
    for i := -1; i < door.Width+1; i++ {
      L.PushInteger(count)
      count++

      L.NewTable()
      L.PushString("x")
      L.PushInteger(room.X + x + dx*i)
      L.SetTable(-3)
      L.PushString("y")
      L.PushInteger(room.Y + y + dy*i)
      L.SetTable(-3)

      L.SetTable(-3)
    }
    return 1
  }
}

// Queries whether a door is currently open.
//    Format
//    open = doorIsOpen(d)
//
//    Input:
//    d - door - A door.
//
//    Output:
//    open - boolean - True if the door is open, false otherwise.
func DoorIsOpenFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "doorIsOpen", luaTable) {
      return 0
    }
    f, r, d := getFloorRoomDoor(L, -1)
    if !checkFloorRoomDoor(a.ent.Game().House, f, r, d) {
      luaDoError(L, fmt.Sprintf("Referenced floor, room, and door (%d, %d, %d) which doesn't exist.", f, r, d))
      return 0
    }
    L.PushBoolean(a.ent.Game().House.Floors[f].Rooms[r].Doors[d].Opened)
    return 1
  }
}

// Performs an Interact action to toggle the opened/closed state of a door.
//    Format
//    res = doDoorToggle(d)
//
//    Input:
//    d - door - A door.
//
//    Output:
//    res - boolean - True if the door was opened, false if it was closed.
//    res will be nil if the action could not be performed for some reason.
func DoDoorToggleFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "doDoorToggle", luaTable) {
      return 0
    }
    f, r, d := getFloorRoomDoor(L, -1)
    if !checkFloorRoomDoor(a.ent.Game().House, f, r, d) {
      luaDoError(L, fmt.Sprintf("Referenced floor, room, and door (%d, %d, %d) which doesn't exist.", f, r, d))
      return 0
    }

    var interact *actions.Interact
    for _, action := range a.ent.Actions {
      var ok bool
      interact, ok = action.(*actions.Interact)
      if ok {
        break
      }
    }
    if interact == nil {
      luaDoError(L, fmt.Sprintf("Tried to toggle a door, but don't have an interact action."))
      L.PushNil()
      return 1
    }
    exec := interact.AiToggleDoor(a.ent, f, r, d)
    if exec != nil {
      a.execs <- exec
      <-a.pause
      L.PushBoolean(a.ent.Game().House.Floors[f].Rooms[r].Doors[d].Opened)
    } else {
      L.PushNil()
    }
    return 1
  }
}

// Returns a list of all positions inside the specified room.
//    Format
//    ps = roomPositions(r)
//
//    Input:
//    r - room - A room.
//
//    Output:
//    ps - array[table[x,y]] - List of all position inside the specified room.
func RoomPositionsFunc(a *Ai) lua.GoFunction {
  return func(L *lua.State) int {
    if !luaCheckParamsOk(L, "roomPositions", luaTable) {
      return 0
    }
    f, r, _ := getFloorRoomDoor(L, -1)
    if !checkFloorRoom(a.ent.Game().House, f, r) {
      luaDoError(L, fmt.Sprintf("Referenced floor and room (%d, %d) which doesn't exist.", f, r))
      return 0
    }
    room := a.ent.Game().House.Floors[f].Rooms[r]

    L.NewTable()
    count := 1
    for x := room.X; x < room.X+room.Size.Dx; x++ {
      for y := room.Y; y < room.Y+room.Size.Dy; y++ {
        L.PushInteger(count)
        count++
        putPointToTable(L, x, y)
        L.SetTable(-3)
      }
    }
    return 1
  }
}
