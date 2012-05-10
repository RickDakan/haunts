package game

import (
  "reflect"
  "math/rand"
  "sort"
  "path/filepath"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/util/algorithm"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/house"
)

type GamePanel struct {
  *gui.AnchorBox

  house  *house.HouseDef
  viewer *house.HouseViewer

  // Only active on turn == 0
  explorer_setup *explorerSetup

  // Only active on turn == 1
  haunt_setup *hauntSetup

  // Only active for turns >= 2
  main_bar *MainBar

  // Keep track of this so we know how much time has passed between
  // calls to Think()
  last_think int64

  complete gui.Widget

  game *Game
}

func MakeGamePanel() *GamePanel {
  var gp GamePanel
  gp.house = house.MakeHouseDef()
  gp.viewer = house.MakeHouseViewer(gp.house, 62)
  gp.AnchorBox = gui.MakeAnchorBox(gui.Dims{1024,700})
  return &gp
}

func (gp *GamePanel) SaveGame(filename string) {
  err := base.SaveGob(filepath.Join(base.GetDataDir(), "games", filename + ".game"), *gp.game)
  if err != nil {
    base.Warn().Printf("Failed to save: %v", err)
  }
}

func (gp *GamePanel) LoadGame(path string) {
  var err error
  gp.game = &Game{}

  err = base.LoadGob(filepath.Join(path), gp.game)
  if err != nil {
    base.Warn().Printf("Failed to load: %v", err)
    return
  }
  base.ProcessObject(reflect.ValueOf(gp.game.House), "")
  base.ProcessObject(reflect.ValueOf(gp.game.Ents), "loadfrom-entities")
  base.ProcessObject(reflect.ValueOf(gp.game.Active_cleanses), "loadfrom-entities")

  gp.game.viewer = house.MakeHouseViewer(gp.house, 62)
  gp.house = gp.game.House
  gp.viewer = gp.game.viewer
  gp.AnchorBox = gui.MakeAnchorBox(gui.Dims{1024,700})
  gp.AnchorBox.AddChild(gp.viewer, gui.Anchor{0.5,0.5,0.5,0.5})

  for i := range gp.game.Ents {
    gp.game.Ents[i].Load(gp.game)
    gp.viewer.AddDrawable(gp.game.Ents[i])
  }

  gp.game.setup()
  gp.game.viewer.Los_tex = gp.game.los_tex
  gp.main_bar,err = MakeMainBar(gp.game)
  if err != nil {
    base.Warn().Printf("Failed to make main bar: %v", err)
    return
  }
  gp.AnchorBox.AddChild(gp.main_bar, gui.Anchor{0.5,0,0.5,0})
}

// Returns  true iff the game panel has an active game with a viewer already
// installed.
func (gp *GamePanel) Active() bool {
  return gp.house != nil && gp.viewer != nil
}

func (gp *GamePanel) Think(ui *gui.Gui, t int64) {
  if !gp.Active() {
    return
  }
  switch gp.game.Turn {
  case 0:
  case 1:
    if gp.explorer_setup != nil {
      gp.AnchorBox.RemoveChild(gp.explorer_setup)
      gp.explorer_setup = nil

      gp.AnchorBox.AddChild(gp.viewer, gui.Anchor{0.5,0.5,0.5,0.5})
      var err error
      gp.haunt_setup,err = MakeHauntSetupBar(gp.game)
      if err == nil {
        gp.AnchorBox.AddChild(gp.haunt_setup, gui.Anchor{0, 0.5, 0, 0.5})
      } else {
        base.Error().Printf("Unable to create haunt setup bar: %v", err)
      }
      // gp.AnchorBox.AddChild(gp.main_bar, gui.Anchor{0.5,0,0.5,0})
    }

  case 2:
    if gp.haunt_setup != nil {
      gp.AnchorBox.RemoveChild(gp.haunt_setup)
      gp.haunt_setup = nil
      gp.AnchorBox.AddChild(gp.main_bar, gui.Anchor{0.5,0,0.5,0})
    }
  default:
  }
  gp.main_bar.SelectEnt(gp.game.selected_ent)
  gp.AnchorBox.Think(ui, t)
  if gp.last_think == 0 {
    gp.last_think = t
  }
  dt := t - gp.last_think
  gp.last_think = t

  gp.game.Think(dt)
  if gp.game.winner != SideNone {
    var name string
    if gp.game.winner == SideExplorers {
      name = "A Winner is an Intruder"
    } else {
      name = "A Winner is a Denizen"
    }
    gp.complete = gui.MakeTextLine("standard", name, 300, 1, 1, 1, 1)
    gp.AnchorBox.AddChild(gp.complete, gui.Anchor{0.5, 0.5, 0.5, 0.5})
  }
}

func (gp *GamePanel) Draw(region gui.Region) {
  gp.AnchorBox.Draw(region)

  region.PushClipPlanes()
  defer region.PopClipPlanes()
}

func (g *Game) SpawnEntity(spawn *Entity, x,y int) {
  spawn.X = float64(x)
  spawn.Y = float64(y)
  g.viewer.AddDrawable(spawn)
  g.Ents = append(g.Ents, spawn)
}

func (g *Game) setupRespond(ui *gui.Gui, group gui.EventGroup) bool {
  if group.Events[0].Key.Id() >= '1' && group.Events[0].Key.Id() <= '9' {
    if group.Events[0].Type == gin.Press {
      index := int(group.Events[0].Key.Id() - '1')
      names := base.GetAllNamesInRegistry("entities")
      ents := algorithm.Map(names, []*Entity{}, func(a interface{}) interface{} {
        return MakeEntity(a.(string), g)
      }).([]*Entity)
      algorithm.Choose2(&ents, func(ent *Entity) bool { return ent.Side() != g.Side })
      if index >= 0 && index < len(ents) {
        g.viewer.RemoveDrawable(g.new_ent)
        g.new_ent = ents[index]
        g.viewer.AddDrawable(g.new_ent)
      }
    }
  }
  return false
}

// Returns true iff the action was set
// This function will return false if there is no selected entity, if the
// action cannot be selected (because it is invalid or the entity has
// insufficient Ap), or if there is an action currently executing.
func (g *Game) SetCurrentAction(action Action) bool {
  if g.action_state != noAction && g.action_state != preppingAction {
    return false
  }
  // the action should be one that belongs to the current entity, if not then
  // we need to bail out immediately
  if g.selected_ent == nil {
    base.Warn().Printf("Tried to SetCurrentAction() without a selected entity.")
    return action == nil
  }
  if action != nil {
    valid := false
    for _, a := range g.selected_ent.Actions {
      if a == action {
        valid = true
        break
      }
    }
    if !valid {
      base.Warn().Printf("Tried to SetCurrentAction() with an action that did not belong to the selected entity.")
      return action == nil
    }
  }
  if g.current_action != nil {
    g.current_action.Cancel()
  }
  if action == nil {
    g.action_state = noAction
  } else {
    g.action_state = preppingAction
  }
  g.current_action = action
  return true
}

func (gp *GamePanel) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if !gp.Active() {
    return false
  }
  if gp.game.winner != SideNone {
    // This is lame - but works for now
    return false
  }
  if gp.AnchorBox.Respond(ui, group) {
    return true
  }

  if gp.game.Turn <= 1 {
    return gp.game.setupRespond(ui, group)
  }

  if group.Events[0].Type == gin.Release {
    return false
  }

  cursor := group.Events[0].Key.Cursor()
  if cursor != nil {
    if gp.game.hovered_ent != nil {
      gp.game.hovered_ent.hovered = false
    }
    gp.game.hovered_ent = nil
    mx,my := cursor.Point()
    for i := range gp.game.Ents {
      fx,fy := gp.game.Ents[i].FPos()
      wx,wy := gp.viewer.BoardToWindow(float32(fx), float32(fy))
      x := wx - int(gp.game.Ents[i].last_render_width / 2)
      y := wy
      x2 := wx + int(gp.game.Ents[i].last_render_width / 2)
      y2 := wy + int(150 * gp.game.Ents[i].last_render_width / 100)
      if mx >= x && mx <= x2 && my >= y && my <= y2 {
        if gp.game.hovered_ent != nil {
          gp.game.hovered_ent.hovered = false
        }
        gp.game.hovered_ent = gp.game.Ents[i]
        gp.game.hovered_ent.hovered = true
      }
    }
  }

  if found,event := group.FindEvent(gin.Escape); found && event.Type == gin.Press {
    base.Log().Printf("Found an escape!: %p %v", gp.game.selected_ent, gp.game.action_state)
    if gp.game.selected_ent != nil {
      switch gp.game.action_state {
      case noAction:
        gp.game.selected_ent.selected = false
        gp.game.selected_ent.hovered = false
        gp.game.selected_ent = nil
        return true

      case preppingAction:
        gp.game.SetCurrentAction(nil)
        return true

      case doingAction:
        // Do nothing - we don't cancel an action that's in progress
      }
    }
  }

  if gp.game.action_state == noAction {
    if found,_ := group.FindEvent(gin.MouseLButton); found {
      if gp.game.hovered_ent != nil && gp.game.hovered_ent.Side() == gp.game.Side {
        if gp.game.selected_ent != nil {
          gp.game.selected_ent.selected = false
        }
        gp.game.selected_ent = gp.game.hovered_ent
        gp.game.selected_ent.selected = true
      }
      return true
    }
  }

  if gp.game.action_state == preppingAction {
    res := gp.game.current_action.HandleInput(group, gp.game)
    switch res {
      case ConsumedAndBegin:
      gp.game.action_state = doingAction
      fallthrough

      case Consumed:
      return true
    }
  }

  // After this point all events that we check for require that we have a
  // selected entity
  if gp.game.selected_ent == nil { return false }
  if gp.game.action_state == noAction || gp.game.action_state == preppingAction {
    if len(group.Events) == 1 && group.Events[0].Key.Id() >= '1' && group.Events[0].Key.Id() <= '9' {
      index := int(group.Events[0].Key.Id() - '1')
      if index >= 0 && index < len(gp.game.selected_ent.Actions) {
        action := gp.game.selected_ent.Actions[index]
        if action != gp.game.current_action && action.Prep(gp.game.selected_ent, gp.game) {
          gp.game.SetCurrentAction(action)
        }
      }
    }
  }

  return false
}

type orderEntsBigToSmall []*Entity
func (o orderEntsBigToSmall) Len() int {
  return len(o)
}
func (o orderEntsBigToSmall) Swap(i, j int) {
  o[i], o[j] = o[j], o[i]
}
func (o orderEntsBigToSmall) Less(i, j int) bool {
  return o[i].Dx * o[i].Dy > o[j].Dx * o[j].Dy
}

type orderSpawnsSmallToBig []*house.SpawnPoint
func (o orderSpawnsSmallToBig) Len() int {
  return len(o)
}
func (o orderSpawnsSmallToBig) Swap(i, j int) {
  o[i], o[j] = o[j], o[i]
}
func (o orderSpawnsSmallToBig) Less(i, j int) bool {
  return o[i].Dx * o[i].Dy < o[j].Dx * o[j].Dy
}

type entSpawnPair struct {
  ent   *Entity
  spawn *house.SpawnPoint
}

// Distributes the ents among the spawn points.  Since this is done randomly
// it might not work, so there is a very small chance that not all spawns will
// have an ent given to them, even if it is possible to distrbiute them
// properly.  Regardless, at least some will be spawned.
func spawnEnts(g *Game, ents []*Entity, spawns []*house.SpawnPoint) {
  sort.Sort(orderSpawnsSmallToBig(spawns))
  sanity := 100
  var places []entSpawnPair
  for sanity > 0 {
    sanity--
    places = places[0:0]
    sort.Sort(orderEntsBigToSmall(ents))
    //slightly shuffle the ents
    for i := range ents {
      j := i + rand.Intn(5) - 2
      if j >= 0 && j < len(ents) {
        ents[i], ents[j] = ents[j], ents[i]
      }
    }
    // Go through each ent and try to place it in an unused spawn point
    used_spawns := make(map[*house.SpawnPoint]bool)
    for _, ent := range ents {
      for _, spawn := range spawns {
        if used_spawns[spawn] { continue }
        if spawn.Dx < ent.Dx || spawn.Dy < ent.Dy { continue }
        used_spawns[spawn] = true
        places = append(places, entSpawnPair{ent, spawn})
        break
      }
    }
    if len(places) == len(spawns) {
      break
    }
  }
  if sanity > 0 {
    base.Log().Printf("Placed all objects with %d sanity remaining", sanity)
  } else {
    base.Warn().Printf("Only able to place %d out of %d objects", len(places), len(spawns))
  }
  for _, place := range places {
    place.ent.X = float64(place.spawn.X + rand.Intn(place.spawn.Dx - place.ent.Dx + 1))
    place.ent.Y = float64(place.spawn.Y + rand.Intn(place.spawn.Dy - place.ent.Dy + 1))
    g.viewer.AddDrawable(place.ent)
    g.Ents = append(g.Ents, place.ent)
    base.Log().Printf("Using object '%s' at (%.0f, %.0f)", place.ent.Name, place.ent.X, place.ent.Y)
  }
}
func spawnAllRelics(g *Game) {
  relic_spawns := algorithm.Choose(g.House.Floors[0].Spawns, func(a interface{}) bool {
    return a.(*house.SpawnPoint).Type() == house.SpawnRelic
  }).([]*house.SpawnPoint)

  ent_names := base.GetAllNamesInRegistry("entities")
  all_ents := algorithm.Map(ent_names, []*Entity{}, func(a interface{}) interface{} {
    return MakeEntity(a.(string), g)
  }).([]*Entity)
  relic_ents := algorithm.Choose(all_ents, func(a interface{}) bool {
    e := a.(*Entity)
    return e.ObjectEnt != nil && e.ObjectEnt.Goal == GoalRelic
  }).([]*Entity)

  spawnEnts(g, relic_ents, relic_spawns)
}
func spawnAllCleanses(g *Game) {
  cleanse_spawns := algorithm.Choose(g.House.Floors[0].Spawns, func(a interface{}) bool {
    return a.(*house.SpawnPoint).Type() == house.SpawnCleanse
  }).([]*house.SpawnPoint)

  ent_names := base.GetAllNamesInRegistry("entities")
  all_ents := algorithm.Map(ent_names, []*Entity{}, func(a interface{}) interface{} {
    return MakeEntity(a.(string), g)
  }).([]*Entity)
  cleanse_ents := algorithm.Choose(all_ents, func(a interface{}) bool {
    e := a.(*Entity)
    return e.ObjectEnt != nil && e.ObjectEnt.Goal == GoalCleanse
  }).([]*Entity)

  spawnEnts(g, cleanse_ents, cleanse_spawns)
}

func (gp *GamePanel) LoadHouse(name string) {
  var err error
  gp.house, err = house.MakeHouseFromPath(name)
  if err != nil {
    base.Error().Printf("%v", err)
    return
  }
  if len(gp.house.Floors) == 0 {
    gp.house = house.MakeHouseDef()
  }
  gp.viewer = house.MakeHouseViewer(gp.house, 62)
  gp.game = makeGame(gp.house, gp.viewer)

  spawnAllRelics(gp.game)
  spawnAllCleanses(gp.game)

  gp.AnchorBox = gui.MakeAnchorBox(gui.Dims{1024,700})

  gp.main_bar,err = MakeMainBar(gp.game)
  if err != nil {
    base.Error().Printf("%v", err)
    panic(err)
  }

  gp.explorer_setup,err = MakeExplorerSetupBar(gp.game)
  if err != nil {
    base.Error().Printf("%v", err)
    panic(err)
  }

  gp.AnchorBox.AddChild(gp.explorer_setup, gui.Anchor{0.5,0.5,0.5,0.5})
}

func (gp *GamePanel) GetViewer() house.Viewer {
  return gp.viewer
}
