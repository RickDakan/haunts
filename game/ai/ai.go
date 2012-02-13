package ai

import (
  "glop/ai"
  "yed"
  "polish"
  "reflect"
  // "haunts/game/actions"
  "haunts/game"
  "haunts/game/actions"
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
}

func makeAi(path string, ent *game.Entity) game.Ai {
  var ai_struct Ai
  ai_graph := ai.NewGraph()
  graph,err := yed.ParseFromFile(path)
  if err != nil {
    panic(err.Error())
  }
  ai_graph.Graph = &graph.Graph
  ai_graph.Context = polish.MakeContext()
  ai_struct.graph = ai_graph
  ai_struct.res = make(chan game.Action)
  ai_struct.pause = make(chan bool)
  ai_struct.ent = ent
  ai_struct.addEntityContext(ai_struct.ent, ai_struct.graph.Context)
  return &ai_struct
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
      println(err.Error())
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

func (a *Ai) addEntityContext(ent *game.Entity, context *polish.Context) {
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

  // Ends an entity's turn
  context.AddFunc("done",
      func() {
        <-a.pause
        // a.graph.Term() <- ai.TermError
      })

  // This entity, the one currently taking its turn
  context.SetValue("me", ent)

  context.AddFunc("move", func() {
    <-a.pause
    var move actions.Move
    x,y := ent.Pos()
    if move.AiMoveToWithin(ent, x-5, y, 1) {
      a.res <- &move
    } else {
      a.graph.Term() <- ai.TermError
    }
  })
}

func numVisibleEntities(e *game.Entity, ally bool) int {
  count := 0
  for _,ent := range e.Game().Ents {
    if ent == e { continue }
    if ent.Stats.HpCur() <= 0 { continue }
    if ally != (e.Side == ent.Side) { continue }
    x,y := ent.Pos()
    if e.HasLos(x, y) {
      count++
    }
  }
  return count
}

func nearestEntity(e *game.Entity, ally bool) *game.Entity {
  var nearest *game.Entity
  cur_dist := 1000000000
  x,y := e.Pos()
  for _,ent := range e.Game().Ents {
    if ent == e { continue }
    if ent.Stats.HpCur() <= 0 { continue }
    if ally != (e.Side == ent.Side) { continue }
    ex,ey := ent.Pos()
    dx := ex - x
    if dx < 0 { dx = -dx }
    dy := ey - y
    if dy < 0 { dy = -dy }
    if cur_dist > dx + dy {
      cur_dist = dx + dy
      nearest = ent
    }
  }
  return nearest
}

// func attack(e *game.Entity, target *game.Entity) {
//   if target == nil {
//     panic("No target")
//   }
//   var att *ActionBasicAttack
//   att = e.getAction(reflect.TypeOf(&ActionBasicAttack{})).(*ActionBasicAttack)
//   if att == nil {
//     panic("couldn't find an attack action")
//   }

//   e.doCmd(func() bool {
//     // TODO: This preamble should be in a level method
//     if att.aiDoAttack(target) {
//       e.level.pending_action = att
//       e.level.mid_action = true
//       return true
//     }
//     return false
//   })
// }

func getAction(e *game.Entity, typ reflect.Type) game.Action {
  for _,action := range e.Actions {
    if reflect.TypeOf(action) == typ {
      return action
    }
  }
  return nil
}

// func doCmd(e *game.Entity, f func() bool) {
//   e.Cmds() <- f
//   cont := <-e.Cont()
//   switch cont {
//     case game.AiEvalCont:

//     case game.AiEvalTerm:
//     e.Ai.Term() <- ai.TermError

//     case game.AiEvalPause:
//     e.Ai.Term() <- ai.InterruptError
//   }
// }

// func advanceTowards(e *game.Entity, target *game.Entity) {
//   if target == nil {
//     panic("No target")
//   }

//   var move *actions.Move
//   move = getAction(e, reflect.TypeOf(&actions.Move{})).(*actions.Move)
//   if move == nil {
//     panic("couldn't find the move action")
//   }

//   doCmd(e, func() bool {
//     // TODO: This preamble should be in a level method
//     tx,ty := target.Pos()
//     if move.AiMoveToWithin(e, tx, ty, 1) {
//       // e.level.pending_action = move
//       // e.level.mid_action = true
//       return true
//     }
//     return false
//   })
// }
