package ai

import (
  "strings"
  "math/rand"
  "reflect"
  "encoding/gob"
  "github.com/runningwild/glop/ai"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  "github.com/runningwild/haunts/game/actions"
  "github.com/runningwild/polish"
  "github.com/runningwild/yedparse"
)

// The Ai struct contains a glop.AiGraph object as well as a few channels for
// communicating with the game itself.
// If there is no action executing and an ent is ready, call Ai.Eval(), it will
// 
type Ai struct {
  graph   *ai.AiGraph

  // Used during evaluation to get the action that the ai wants to execute
  res chan game.Action

  // Once we send an Action for execution we have to wait until it is done
  // before we make the next one.  This channel is used to handle that.
  pause chan bool

  // Keep track of the ent since we'll want to reference it regularly
  ent *game.Entity

  State AiState
}

type AiState struct {
  Last_offensive_target game.EntityId
}

func init() {
  gob.Register(&Ai{})
}

func makeAi(path string, ent *game.Entity, dst_iface *game.Ai) {
  var ai_struct *Ai
  if *dst_iface == nil {
    ai_struct = new(Ai)
  } else {
    ai_struct = (*dst_iface).(*Ai)
  }
  ai_graph := ai.NewGraph()
  graph,err := yed.ParseFromFile(path)
  if err != nil {
    base.Error().Printf("%v", err)
    panic(err.Error())
  }
  ai_graph.Graph = &graph.Graph
  ai_graph.Context = polish.MakeContext()
  ai_struct.graph = ai_graph
  ai_struct.res = make(chan game.Action)
  ai_struct.pause = make(chan bool)
  ai_struct.ent = ent
  ai_struct.addEntityContext(ai_struct.ent, ai_struct.graph.Context)
  *dst_iface = ai_struct
}

func init() {
  game.SetAiMaker(makeAi)
}

// Eval() evaluates the ai graph and returns an Action that the entity wants
// to execute, or nil if the entity is done for the turn.
func (a *Ai) Eval() {
  go func() {
    err := a.graph.Eval()
    a.res <- nil
    if err != nil {
      base.Warn().Printf("%v", err)
    }
  } ()
}

func (a *Ai) Actions() <-chan game.Action {
  select {
    case a.pause <- true:
    default:
  }
  return a.res
}

// Does the roll dice-d-sides, like 3d6, and returns the result
func roll(dice, sides int) int {
  result := 0
  for i := 0; i < dice; i++ {
    result += rand.Intn(sides) + 1
  }
  return result
}

func (a *Ai) addEntityContext(ent *game.Entity, context *polish.Context) {
  polish.AddIntMathContext(context)

  // This entity, the one currently taking its turn
  context.SetValue("me", ent)

  // All actions that the entity has are available using their names,
  // converted to lower case, and replacing spaces with underscores.
  // For example, "Kiss of Death" -> "kiss_of_death"
  for _, action := range ent.Actions {
    name := lowerAndUnderscore(action.String())
    context.SetValue(name, action)
  }

  // rolls dice, for example "roll 3 6" is a roll of 3d6
  context.AddFunc("roll", roll)

  // These functions are self-explanitory, they are all relative to the
  // current entity
  context.AddFunc("numVisibleEnemies",
      func() int {
        return numVisibleEntities(ent, false)
      })
  context.AddFunc("nearestEnemy",
      func() *game.Entity {
        return nearestEntity(ent, false)
      })
  context.AddFunc("distBetween", distBetween)

  // Ends an entity's turn
  context.AddFunc("done",
      func() {
        <-a.pause
      })

  // Checks whether an entity is nil, this is important to check when using
  // function that returns an entity (like lastOffensiveTarget)
  context.AddFunc("stillExists", func(target *game.Entity) bool {
    return target != nil
  })

  // Returns the last entity that this ai attacked.  If the entity has died
  // this can return nil, so be sure to check that before using it.
  context.AddFunc("lastOffensiveTarget", func() *game.Entity {
    return ent.Game().EntityById(a.State.Last_offensive_target)
  })

  // Advances as far as possible towards the target entity.
  context.AddFunc("advanceTowards", func(target *game.Entity) {
    move := getAction(ent, reflect.TypeOf(&actions.Move{})).(*actions.Move)
    x,y := target.Pos()
    if move.AiMoveToWithin(ent, x, y, 1) {
      a.res <- move
    } else {
      a.graph.Term() <- ai.TermError
    }
    <-a.pause
    x,y = ent.Pos()
  })

  context.AddFunc("getBasicAttack", func() game.Action {
    return getAction(ent, reflect.TypeOf(&actions.BasicAttack{})).(*actions.BasicAttack)
  })

  context.AddFunc("doBasicAttack", func(target *game.Entity, _attack game.Action) {
    attack := _attack.(*actions.BasicAttack)
    if attack.AiAttackTarget(ent, target) {
      base.Log().Printf("Ent(%p) attacking (%p) with %v", ent, target, attack)
      a.res <- attack
      a.State.Last_offensive_target = target.Id
    } else {
      a.graph.Term() <- ai.TermError
    }
    <-a.pause
  })
}

func numVisibleEntities(e *game.Entity, ally bool) int {
  count := 0
  for _,ent := range e.Game().Ents {
    if ent == e { continue }
    if ent.Stats == nil || ent.Stats.HpCur() <= 0 { continue }
    if ally != (e.Side() == ent.Side()) { continue }
    x,y := ent.Pos()
    if e.HasLos(x, y, 1, 1) {
      count++
    }
  }
  return count
}

func distBetween(e1,e2 *game.Entity) int {
  e1x,e1y := e1.Pos()
  e2x,e2y := e2.Pos()
  dx := e1x - e2x
  dy := e1y - e2y
  if dx < 0 { dx = -dx }
  if dy < 0 { dy = -dy }
  if dx > dy {
    return dx
  }
  return dy
}

func nearestEntity(e *game.Entity, ally bool) *game.Entity {
  var nearest *game.Entity
  cur_dist := 1000000000
  for _,ent := range e.Game().Ents {
    if ent == e { continue }
    if ent.Stats == nil || ent.Stats.HpCur() <= 0 { continue }
    if ally != (e.Side() == ent.Side()) { continue }
    dist := distBetween(e, ent)
    if cur_dist > dist {
      cur_dist = dist
      nearest = ent
    }
  }
  return nearest
}

func lowerAndUnderscore(s string) string {
  b := []byte(strings.ToLower(s))
  for i := range b {
    if b[i] == ' ' {
      b[i] = '_'
    }
  }
  return string(b)
}

func getActionByName(e *game.Entity, name string) game.Action {
  for _,action := range e.Actions {
    if lowerAndUnderscore(action.String()) == name {
      return action
    }
  }
  return nil
}

func getAction(e *game.Entity, typ reflect.Type) game.Action {
  for _,action := range e.Actions {
    if reflect.TypeOf(action) == typ {
      return action
    }
  }
  return nil
}
