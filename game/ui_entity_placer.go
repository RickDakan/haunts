package game

import (
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/house"
  "github.com/runningwild/haunts/game/hui"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/util/algorithm"
)

type entityPlacer struct {
  *gui.AnchorBox
  roster_chooser *hui.RosterChooser
  game           *Game
  points         int

  names        []string
  name_to_cost map[string]int

  // regular expression that matches valid spawn points
  pattern string

  undo_stack []*Entity
}

func makeEntityPlacerSelector(game *Game, ep *entityPlacer) hui.Selector {
  return func(index int, selected map[int]bool, doit bool) (valid bool) {
    if index == -1 {
      return true
    } else {
      valid = ep.name_to_cost[ep.names[index]] <= ep.points
      if !valid {
        delete(selected, index)
        return
      }
    }
    if doit {
      var other int
      for k, _ := range selected {
        other = k
      }
      delete(selected, other)
      selected[index] = true
      if game.new_ent == nil || game.new_ent.Name != ep.names[index] {
        game.viewer.RemoveDrawable(game.new_ent)
        game.new_ent = MakeEntity(ep.names[index], game)
        game.viewer.AddDrawable(game.new_ent)
      }
    }
    return
  }
}

func getAllEntsWithSideAndLevel(game *Game, side Side, level EntLevel) []*Entity {
  names := base.GetAllNamesInRegistry("entities")
  ents := algorithm.Map(names, []*Entity{}, func(a interface{}) interface{} {
    return MakeEntity(a.(string), game)
  }).([]*Entity)
  ents = algorithm.Choose(ents, func(a interface{}) bool {
    return a.(*Entity).Side() == side && a.(*Entity).HauntEnt.Level == level
  }).([]*Entity)
  return ents
}

// Returns an entity placer and a channel that will send a single slice
// containing the entities that were placed.
func MakeEntityPlacer(g *Game, pattern string, points int, names []string, costs []int) (*entityPlacer, <-chan []*Entity) {
  house.PushSpawnRegexp(pattern)
  ep := entityPlacer{
    game:         g,
    points:       points,
    names:        names,
    name_to_cost: make(map[string]int),
    pattern:      pattern,
  }

  var roster []hui.Option
  for i := range ep.names {
    ep.name_to_cost[ep.names[i]] = costs[i]
    roster = append(roster, makeEntLabel(MakeEntity(ep.names[i], g)))
  }

  placed := make(chan []*Entity, 1)
  ep.roster_chooser = hui.MakeRosterChooser(roster,
    makeEntityPlacerSelector(g, &ep),
    func(m map[int]bool) {
      placed <- ep.undo_stack
      house.PopSpawnRegexp()
      close(placed)
    },
    func() {
      if len(ep.undo_stack) > 0 {
        top_pos := len(ep.undo_stack) - 1
        top := ep.undo_stack[top_pos]
        ep.undo_stack = ep.undo_stack[0:top_pos]
        g.GetViewer().RemoveDrawable(top)
        algorithm.Choose2(&g.Ents, func(e *Entity) bool { return e != top })
        ep.points += ep.name_to_cost[top.Name]
      }
    },
  )
  ep.AnchorBox = gui.MakeAnchorBox(gui.Dims{1024, 768})
  ep.AnchorBox.AddChild(ep.roster_chooser, gui.Anchor{0, 0.5, 0, 0.5})

  return &ep, placed
}

func (ep *entityPlacer) Think(ui *gui.Gui, t int64) {
  ep.AnchorBox.Think(ui, t)
}

func (ep *entityPlacer) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if ep.AnchorBox.Respond(ui, group) {
    return true
  }
  if ep.game.new_ent != nil {
    x, y := gin.In().GetCursor("Mouse").Point()
    fbx, fby := ep.game.viewer.WindowToBoard(x, y)
    bx, by := DiscretizePoint32(fbx, fby)
    ep.game.new_ent.X, ep.game.new_ent.Y = float64(bx), float64(by)
    if found, event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
      ent := ep.game.new_ent
      if ep.game.placeEntity(ep.pattern) {
        cost := ep.name_to_cost[ent.Name]
        ep.points -= cost
        ep.undo_stack = append(ep.undo_stack, ent)
        if cost <= ep.points {
          ep.game.new_ent = MakeEntity(ent.Name, ep.game)
          ep.game.viewer.AddDrawable(ep.game.new_ent)
        }
      }
      return true
    }
  }
  return false
}

func (ep *entityPlacer) String() string {
  return "entity placer"
}
