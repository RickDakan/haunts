package game

import (
  // "path/filepath"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/util/algorithm"
  "github.com/runningwild/haunts/house"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/cmwc"
)

type Purpose int

const (
  PurposeNone Purpose = iota
  PurposeRelic
  PurposeMystery
  PurposeCleanse
)

type LosMode int

const (
  LosModeNone LosMode = iota
  LosModeBlind
  LosModeAll
  LosModeEntities
  LosModeRooms
)

type turnState int

const (
  // Waiting for the script to finish Init()
  turnStateInit turnState = iota

  // Waiting for the script to finish RoundStart()
  turnStateStart

  // Waiting for or running an Ai action
  turnStateAiAction

  // Waiting for the script to finish OnAction()
  turnStateScriptOnAction

  // Humans and Ai are done, now the script can run some actions if it wants
  turnStateMainPhaseOver

  // Waiting for the script to finish OnEnd()
  turnStateEnd
)

type sideLosData struct {
  mode LosMode
  tex  *house.LosTexture
}

type Game struct {
  // TODO: No idea if this thing can be loaded from the registry - should
  // probably figure that out at some point
  House *house.HouseDef
  Ents  []*Entity

  // Set of all Entities that are still resident.  This is so we can safely
  // clean things up since they will all have ais running in the background
  // preventing them from getting GCed.
  all_ents_in_game   map[*Entity]bool
  all_ents_in_memory map[*Entity]bool

  // Next unique EntityId to be assigned
  Entity_id EntityId

  // Current player
  Side Side

  // Current turn number - incremented on each OnRound() so every two
  // indicates that a complete round has happened.
  Turn int

  // PRNG, need it here so that we serialize it along with everything
  // else so that replays work properly.
  Rand *cmwc.Cmwc

  // The active cleanse points - when interacted with they will be removed
  // from this list, so in a Cleanse scenario the mission is accomplished
  // when this list is empty.  In other scenarios this list is always empty.
  Active_cleanses []*Entity

  // The active relics - when interacted it will be nilified
  // If the scenario is not a Relic mission this is always nil
  Active_relic *Entity

  // Transient data - none of the following are exported

  player_active bool

  viewer *house.HouseViewer

  selection_tab      *gui.TabFrame
  haunts_selection   *gui.VerticalTable
  explorer_selection *gui.VerticalTable

  // If the user is dragging around a new Entity to place, this is it
  new_ent *Entity

  los struct {
    denizens, intruders sideLosData

    // When merging the los from different entities we'll do it here, and we
    // keep it around to avoid reallocating it every time we need it.
    full_merger []bool
    merger      [][]bool
  }

  selected_ent *Entity
  hovered_ent  *Entity

  // Stores the current acting entity - if it is an Ai controlled entity
  ai_ent *Entity

  // This controls the minions on the haunting side, regardless of whether a
  // human or ai is controlling the rest of that side.
  minion_ai Ai

  haunts_ai    Ai
  explorers_ai Ai

  // If an Ai is executing currently it is referenced here
  active_ai Ai

  action_state   actionState
  current_exec   ActionExec
  current_action Action

  // Indicates if we're waiting for a script to run or something
  turn_state turnState

  // Used to sync up with the script, the value passed is usually nil, but
  // whenever an action happens it will get passed along this channel too.
  comm struct {
    script_to_game chan interface{}
    game_to_script chan interface{}
  }

  // Goals ******************************************************

  // Defaults to the zero value which is NoSide
  winner Side

  script *gameScript
}

func (g *Game) EntityById(id EntityId) *Entity {
  for i := range g.Ents {
    if g.Ents[i].Id == id {
      return g.Ents[i]
    }
  }
  return nil
}

func (g *Game) HoveredEnt() *Entity {
  return g.hovered_ent
}

func (g *Game) SelectEnt(ent *Entity) bool {
  if g.action_state != noAction {
    return false
  }
  found := false
  for i := range g.Ents {
    if g.Ents[i] == ent {
      found = true
      break
    }
  }
  if !found {
    return false
  }
  if g.selected_ent != nil {
    g.selected_ent.selected = false
    g.selected_ent.hovered = false
  }
  g.selected_ent = ent
  if g.selected_ent != nil {
    g.selected_ent.selected = true
  }
  g.viewer.Focus(ent.FPos())
  return true
}

func (g *Game) OnBegin() {
  for i := range g.Ents {
    if g.Ents[i].Stats != nil {
      g.Ents[i].Stats.OnBegin()
    }
  }
}

// TODO: DEPRECATED
func (g *Game) PlaceInitialExplorers(ents []*Entity) {
}

func (g *Game) checkWinConditions() {
  return
  // Check for explorer win conditions
  explorer_win := false

  if explorer_win {
    base.Log().Printf("Explorers won - kaboom")
  }

  // Check for haunt win condition - all intruders dead
  haunts_win := true
  for i := range g.Ents {
    if g.Ents[i].Side() == SideExplorers {
      haunts_win = false
    }
  }
  if haunts_win {
    base.Log().Printf("Haunts won - kaboom")
  }

  if g.winner == SideNone {
    if explorer_win && !haunts_win {
      g.winner = SideExplorers
    }
    if haunts_win && !explorer_win {
      g.winner = SideHaunt
    }
    if explorer_win && haunts_win {
      // Let's just hope this is far beyond impossible
      base.Error().Printf("Both sides won at the same time omg!")
    }
  }
}

func (g *Game) SetVisibility(side Side) {
  switch side {
  case SideHaunt:
    g.viewer.Los_tex = g.los.denizens.tex
  case SideExplorers:
    g.viewer.Los_tex = g.los.intruders.tex
  default:
    base.Error().Printf("Unable to SetVisibility for side == %d.", side)
    return
  }
}

// This is called if the player is ready to end the turn, if the turn ends
// then the following things happen:
// 1. The game script gets to run its OnRound() function
// 2. Entities with stats and HpCur() <= 0 are removed.
// 3. Entities all have their OnRound() function called.
func (g *Game) OnRound() {
  // Don't end the round if any of the following are true
  // An action is currently executing
  if g.action_state != noAction {
    return
  }
  // Any master ai is still active
  if g.Side == SideHaunt && (g.minion_ai.Active() || g.haunts_ai.Active()) {
    return
  }

  g.Turn++
  if g.Side == SideExplorers {
    g.Side = SideHaunt
  } else {
    g.Side = SideExplorers
  }
  g.viewer.Los_tex.Remap()

  for i := range g.Ents {
    if g.Ents[i].Side() == g.Side {
      g.Ents[i].OnRound()
    }
  }

  // The entity ais must be activated before the master ais, otherwise the
  // masters might be running with stale data if one of the entities has been
  // reloaded.
  for i := range g.Ents {
    g.Ents[i].Ai.Activate()
    base.Log().Printf("EntityActive '%s': %t", g.Ents[i].Name, g.Ents[i].Ai.Active())
  }

  if g.Side == SideHaunt {
    g.minion_ai.Activate()
    base.Log().Printf("minion ai: %t", g.minion_ai.Active())
    g.haunts_ai.Activate()
    g.player_active = !g.haunts_ai.Active()
  } else {
    g.explorers_ai.Activate()
    g.player_active = !g.explorers_ai.Active()
  }

  for i := range g.Ents {
    if g.Ents[i].Stats != nil && g.Ents[i].Stats.HpCur() <= 0 {
      g.viewer.RemoveDrawable(g.Ents[i])
    }
  }
  algorithm.Choose2(&g.Ents, func(ent *Entity) bool {
    return ent.Stats == nil || ent.Stats.HpCur() > 0
  })

  g.script.OnRound(g)

  if g.selected_ent != nil {
    g.selected_ent.hovered = false
    g.selected_ent.selected = false
  }
  if g.hovered_ent != nil {
    g.hovered_ent.hovered = false
    g.hovered_ent.selected = false
  }
  g.hovered_ent = nil
}

type actionState int

const (
  noAction actionState = iota

  // The Ai is running and determining the next action to run
  waitingAction

  // The player has selected an action and is determining whether or not to
  // use it, and how.
  preppingAction

  // Check the scripts to see if the action should be modified or cancelled.
  verifyingAction

  // An action is currently running, everything should pause while this runs.
  doingAction
)

func (g *Game) GetViewer() *house.HouseViewer {
  return g.viewer
}

func (g *Game) numVertex() int {
  total := 0
  for _, room := range g.House.Floors[0].Rooms {
    total += room.Size.Dx * room.Size.Dy
  }
  return total
}
func (g *Game) FromVertex(v int) (room *house.Room, x, y int) {
  for _, room := range g.House.Floors[0].Rooms {
    size := room.Size.Dx * room.Size.Dy
    if v >= size {
      v -= size
      continue
    }
    return room, room.X + (v % room.Size.Dx), room.Y + (v / room.Size.Dx)
  }
  return nil, 0, 0
}
func (g *Game) ToVertex(x, y int) int {
  v := 0
  for _, room := range g.House.Floors[0].Rooms {
    if x >= room.X && y >= room.Y && x < room.X+room.Size.Dx && y < room.Y+room.Size.Dy {
      x -= room.X
      y -= room.Y
      v += x + y*room.Size.Dx
      break
    }
    v += room.Size.Dx * room.Size.Dy
  }
  return v
}

// x and y are given in room coordinates
func furnitureAt(room *house.Room, x, y int) *house.Furniture {
  for _, f := range room.Furniture {
    fx, fy := f.Pos()
    fdx, fdy := f.Dims()
    if x >= fx && x < fx+fdx && y >= fy && y < fy+fdy {
      return f
    }
  }
  return nil
}

// x and y are given in floor coordinates
func roomAt(floor *house.Floor, x, y int) *house.Room {
  for _, room := range floor.Rooms {
    rx, ry := room.Pos()
    rdx, rdy := room.Dims()
    if x >= rx && x < rx+rdx && y >= ry && y < ry+rdy {
      return room
    }
  }
  return nil
}

func connected(r, r2 *house.Room, x, y, x2, y2 int) bool {
  if r == r2 {
    return true
  }
  x -= r.X
  y -= r.Y
  x2 -= r2.X
  y2 -= r2.Y
  var facing house.WallFacing
  if x == 0 && x2 != 0 {
    facing = house.NearLeft
  } else if y == 0 && y2 != 0 {
    facing = house.NearRight
  } else if x != 0 && x2 == 0 {
    facing = house.FarRight
  } else if y != 0 && y2 == 0 {
    facing = house.FarLeft
  } else {
    // This shouldn't happen, but in case it does we certainly shouldn't treat
    // it as an open door
    return false
  }
  for _, door := range r.Doors {
    if door.Facing != facing {
      continue
    }
    var pos int
    switch facing {
    case house.NearLeft:
      fallthrough
    case house.FarRight:
      pos = y

    case house.NearRight:
      fallthrough
    case house.FarLeft:
      pos = x
    }
    if pos >= door.Pos && pos < door.Pos+door.Width {
      return door.IsOpened()
    }
  }
  return false
}

func (g *Game) IsCellOccupied(x, y int) bool {
  r := roomAt(g.House.Floors[0], x, y)
  if r == nil {
    return true
  }
  f := furnitureAt(r, x-r.X, y-r.Y)
  if f != nil {
    return true
  }
  for _, ent := range g.Ents {
    ex, ey := ent.Pos()
    if x == ex && y == ey {
      return true
    }
  }
  return false
}

type exclusionGraph struct {
  side Side
  ex   map[*Entity]bool
  g    *Game
}

func (eg *exclusionGraph) Adjacent(v int) ([]int, []float64) {
  return eg.g.adjacent(v, eg.side, eg.ex)
}
func (eg *exclusionGraph) NumVertex() int {
  return eg.g.numVertex()
}

func (g *Game) RecalcLos() {
  for i := range g.Ents {
    if g.Ents[i].los != nil {
      g.Ents[i].los.x = -1
    }
  }
}

type roomGraph struct {
  g *Game
}

func (g *Game) RoomGraph() algorithm.Graph {
  return &roomGraph{g}
}

func (rg *roomGraph) NumVertex() int {
  return len(rg.g.House.Floors[0].Rooms)
}

func (rg *roomGraph) Adjacent(n int) ([]int, []float64) {
  room := rg.g.House.Floors[0].Rooms[n]
  var adj []int
  var cost []float64
  for _, door := range room.Doors {
    other_room, _ := rg.g.House.Floors[0].FindMatchingDoor(room, door)
    if other_room != nil {
      for i := range rg.g.House.Floors[0].Rooms {
        if other_room == rg.g.House.Floors[0].Rooms[i] {
          adj = append(adj, i)
          cost = append(cost, 1)
          break
        }
      }
    }
  }
  return adj, cost
}

func (g *Game) Graph(side Side, exclude []*Entity) algorithm.Graph {
  ex := make(map[*Entity]bool, len(exclude))
  for i := range exclude {
    ex[exclude[i]] = true
  }
  return &exclusionGraph{side, ex, g}
}

func (g *Game) adjacent(v int, side Side, ex map[*Entity]bool) ([]int, []float64) {
  room, x, y := g.FromVertex(v)
  var adj []int
  var weight []float64
  var moves [3][3]float64
  ent_occupied := make(map[[2]int]bool)
  for _, ent := range g.Ents {
    if ex[ent] {
      continue
    }
    x, y := ent.Pos()
    dx, dy := ent.Dims()
    for i := x; i < x+dx; i++ {
      for j := y; j < y+dy; j++ {
        ent_occupied[[2]int{i, j}] = true
      }
    }
  }
  var data *sideLosData
  switch side {
  case SideHaunt:
    data = &g.los.denizens
  case SideExplorers:
    data = &g.los.intruders
  default:
    base.Error().Printf("Unable to SetLosMode for side == %d.", side)
    return nil, nil
  }
  for dx := -1; dx <= 1; dx++ {
    for dy := -1; dy <= 1; dy++ {
      // Only run this loop if exactly one of dx and dy is non-zero
      if (dx == 0) == (dy == 0) {
        continue
      }
      tx := x + dx
      ty := y + dy
      if ent_occupied[[2]int{tx, ty}] {
        continue
      }
      if data.tex.Pix()[tx][ty] < house.LosVisibilityThreshold {
        continue
      }
      // TODO: This is obviously inefficient
      troom, _, _ := g.FromVertex(g.ToVertex(tx, ty))
      if troom == nil {
        continue
      }
      if furnitureAt(troom, tx-troom.X, ty-troom.Y) != nil {
        continue
      }
      if !connected(room, troom, x, y, tx, ty) {
        continue
      }
      adj = append(adj, g.ToVertex(tx, ty))
      moves[dx+1][dy+1] = 1
      weight = append(weight, 1)
    }
  }
  for dx := -1; dx <= 1; dx++ {
    for dy := -1; dy <= 1; dy++ {
      // Only run this loop if both dx and dy are non-zero (moving diagonal)
      if (dx == 0) != (dy == 0) {
        continue
      }
      tx := x + dx
      ty := y + dy
      if ent_occupied[[2]int{tx, ty}] {
        continue
      }
      if data.tex.Pix()[tx][ty] < house.LosVisibilityThreshold {
        continue
      }
      // TODO: This is obviously inefficient
      troom, _, _ := g.FromVertex(g.ToVertex(tx, ty))
      if troom == nil {
        continue
      }
      if furnitureAt(troom, tx-troom.X, ty-troom.Y) != nil {
        continue
      }
      if !connected(room, troom, x, y, tx, ty) {
        continue
      }
      if !connected(troom, room, tx, ty, x, y) {
        continue
      }
      if moves[dx+1][1] == 0 || moves[1][dy+1] == 0 {
        continue
      }
      adj = append(adj, g.ToVertex(tx, ty))
      w := (moves[dx+1][1] + moves[1][dy+1]) / 2
      moves[dx+1][dy+1] = w
      weight = append(weight, w)
    }
  }
  return adj, weight
}

func (g *Game) setup() {
  g.los.denizens.tex = house.MakeLosTexture()
  g.los.intruders.tex = house.MakeLosTexture()
  g.los.full_merger = make([]bool, house.LosTextureSizeSquared)
  g.los.merger = make([][]bool, house.LosTextureSize)
  for i := range g.los.merger {
    g.los.merger[i] = g.los.full_merger[i*house.LosTextureSize : (i+1)*house.LosTextureSize]
  }
  for i := range g.Ents {
    if g.Ents[i].Side() == g.Side {
      g.UpdateEntLos(g.Ents[i], true)
    }
  }
  g.viewer.Los_tex = g.los.denizens.tex

  g.minion_ai = inactiveAi{}
  g.haunts_ai = inactiveAi{}
  g.explorers_ai = inactiveAi{}
}

func makeGame(h *house.HouseDef, viewer *house.HouseViewer, side Side) *Game {
  var g Game
  g.Side = SideExplorers
  g.House = h
  g.House.Normalize()
  g.viewer = viewer
  g.Rand = cmwc.MakeCmwc(4285415527, 3)
  g.Rand.SeedWithDevRand()
  g.all_ents_in_game = make(map[*Entity]bool)
  g.all_ents_in_memory = make(map[*Entity]bool)

  // This way an unset id will be invalid
  g.Entity_id = 1

  g.setup()

  g.explorer_selection = gui.MakeVerticalTable()
  g.explorer_selection.AddChild(gui.MakeTextLine("standard", "foo", 300, 1, 1, 1, 1))
  g.haunts_selection = gui.MakeVerticalTable()
  g.haunts_selection.AddChild(gui.MakeTextLine("standard", "bar", 300, 1, 1, 1, 1))
  g.selection_tab = gui.MakeTabFrame([]gui.Widget{g.explorer_selection, g.haunts_selection})

  g.comm.script_to_game = make(chan interface{}, 1)
  g.comm.game_to_script = make(chan interface{}, 1)

  return &g
}

func (g *Game) SetLosMode(side Side, mode LosMode, rooms []*house.Room) {
  var data *sideLosData
  switch side {
  case SideHaunt:
    data = &g.los.denizens
  case SideExplorers:
    data = &g.los.intruders
  default:
    base.Error().Printf("Unable to SetLosMode for side == %d.", side)
    return
  }
  data.mode = mode
  pix := data.tex.Pix()

  switch data.mode {
  case LosModeNone:
    for i := range pix {
      for j := range pix[i] {
        if pix[i][j] >= house.LosVisibilityThreshold {
          pix[i][j] = house.LosVisibilityThreshold - 1
        }
      }
    }

  case LosModeBlind:
    for i := range pix {
      for j := range pix[i] {
        if pix[i][j] >= house.LosVisibilityThreshold {
          pix[i][j] = 0
        }
      }
    }

  case LosModeAll:
    for i := range pix {
      for j := range pix[i] {
        if pix[i][j] < house.LosVisibilityThreshold {
          pix[i][j] = house.LosVisibilityThreshold
        }
      }
    }

  case LosModeEntities:
    // Don't need to do anything here - it's handled on every think

  case LosModeRooms:
    in_room := make(map[int]bool)
    for _, room := range rooms {
      for x := room.X; x < room.X+room.Size.Dx; x++ {
        for y := room.Y; y < room.Y+room.Size.Dy; y++ {
          in_room[g.ToVertex(x, y)] = true
        }
      }
    }
    for i := range pix {
      for j := range pix[i] {
        if in_room[g.ToVertex(i, j)] {
          if pix[i][j] < house.LosVisibilityThreshold {
            pix[i][j] = house.LosVisibilityThreshold
          }
        } else {
          if pix[i][j] >= house.LosVisibilityThreshold {
            pix[i][j] = house.LosVisibilityThreshold - 1
          }
        }

      }
    }
  }
  data.tex.Remap()
}

func (g *Game) Think(dt int64) {

  for _, ent := range g.Ents {
    if !g.all_ents_in_game[ent] {
      g.all_ents_in_game[ent] = true
      g.all_ents_in_memory[ent] = true
    }
  }
  var mark []*Entity
  for ent := range g.all_ents_in_memory {
    if !g.all_ents_in_game[ent] && ent != g.new_ent {
      mark = append(mark, ent)
    }
  }
  for _, ent := range mark {
    delete(g.all_ents_in_game, ent)
    delete(g.all_ents_in_memory, ent)
    ent.Release()
  }

  switch g.turn_state {
  case turnStateInit:
    select {
    case <-g.comm.script_to_game:
      base.Log().Printf("ScriptComm: change to turnStateStart")
      g.turn_state = turnStateStart
      g.OnRound()
    default:
    }
  case turnStateStart:
    select {
    case <-g.comm.script_to_game:
      base.Log().Printf("ScriptComm: change to turnStateAiAction")
      g.turn_state = turnStateAiAction
    default:
    }
  case turnStateScriptOnAction:
    select {
    case exec := <-g.comm.script_to_game:
      if g.current_exec != nil {
        base.Error().Printf("Got an exec from the script when we already had one pending.")
      } else {
        if exec != nil {
          g.current_exec = exec.(ActionExec)
        } else {
          base.Log().Printf("ScriptComm: change to turnStateAiAction")
          g.turn_state = turnStateAiAction
        }
      }
    default:
    }
  case turnStateMainPhaseOver:
    select {
    case exec := <-g.comm.script_to_game:
      if exec != nil {
        base.Log().Printf("ScriptComm: Got an exec: %v", exec)
        g.current_exec = exec.(ActionExec)
        g.action_state = doingAction
        ent := g.EntityById(g.current_exec.EntityId())
        g.current_action = ent.Actions[g.current_exec.ActionIndex()]
        ent.current_action = g.current_action
      } else {
        g.turn_state = turnStateEnd
        base.Log().Printf("ScriptComm: change to turnStateEnd for realzes")
      }
    default:
      base.Log().Printf("ScriptComm: turnStateMainPhaseOver default")
    }

  case turnStateEnd:
    select {
    case <-g.comm.script_to_game:
      g.turn_state = turnStateStart
      base.Log().Printf("ScriptComm: change to turnStateStart")
      g.OnRound()
    default:
    }
  }

  if g.current_exec != nil && g.action_state != verifyingAction && g.turn_state != turnStateMainPhaseOver {
    ent := g.EntityById(g.current_exec.EntityId())
    g.current_action = ent.Actions[g.current_exec.ActionIndex()]
    ent.current_action = g.current_action
    g.action_state = verifyingAction
    base.Log().Printf("ScriptComm: request exec verification")
    g.comm.game_to_script <- g.current_exec
  }

  if g.action_state == verifyingAction {
    select {
    case <-g.comm.script_to_game:
      g.action_state = doingAction
    default:
    }
  }

  // If there is an action that is currently executing we need to advance that
  // action.
  if g.action_state == doingAction {
    res := g.current_action.Maintain(dt, g, g.current_exec)
    if g.current_exec != nil {
      base.Log().Printf("ScriptComm: sent action")
      g.current_exec = nil
    }
    switch res {
    case Complete:
      g.current_action.Cancel()
      g.current_action = nil
      g.action_state = noAction
      if g.turn_state != turnStateMainPhaseOver {
        g.turn_state = turnStateScriptOnAction
      }
      base.Log().Printf("ScriptComm: Action complete")
      g.comm.game_to_script <- nil
      g.checkWinConditions()

    case InProgress:
    case CheckForInterrupts:
    }
  }

  g.viewer.Floor_drawer = g.current_action
  for _, ent := range g.Ents {
    ent.Think(dt)
    s := ent.Sprite()
    if s.AnimState() == "ready" && s.Idle() && g.current_action == nil && ent.current_action != nil {
      ent.current_action = nil
    }
  }
  if g.new_ent != nil {
    g.new_ent.Think(dt)
  }

  for i := range g.Ents {
    g.UpdateEntLos(g.Ents[i], false)
  }
  if g.los.denizens.mode == LosModeEntities {
    g.mergeLos(SideHaunt)
  }
  if g.los.intruders.mode == LosModeEntities {
    g.mergeLos(SideExplorers)
  }

  for _, tex := range []*house.LosTexture{g.los.denizens.tex, g.los.intruders.tex} {
    pix := tex.Pix()
    amt := dt/6 + 1
    mod := false
    for i := range pix {
      for j := range pix[i] {
        v := int64(pix[i][j])
        if v < house.LosVisibilityThreshold {
          v -= amt
        } else {
          v += amt
        }
        if v < house.LosMinVisibility {
          v = house.LosMinVisibility
        }
        if v < 0 {
          v = 0
        }
        if v > 255 {
          v = 255
        }
        mod = mod || (byte(v) != pix[i][j])
        pix[i][j] = byte(v)
      }
    }
    if mod {
      tex.Remap()
    }
  }

  // Don't do any ai stuff if there is a pending action
  if g.current_action != nil {
    return
  }

  // Also don't do an ai stuff if this isn't the appropriate state
  if g.turn_state != turnStateAiAction {
    return
  }

  // If any entities are not either ready or dead let's wait until they are
  // before we do any of the ai stuff
  for _, ent := range g.Ents {
    if ent.Side() != SideHaunt && ent.Side() != SideExplorers {
      // Relics and cleanse points and whatnot matter here, and they might not
      // be in a 'ready' state.
      continue
    }
    state := ent.sprite.Sprite().AnimState()
    if state != "ready" && state != "killed" {
      return
    }
    if !ent.sprite.Sprite().Idle() {
      return
    }
  }

  // Do Ai - if there is any to do
  if g.Side == SideHaunt {
    if g.minion_ai.Active() {
      g.active_ai = g.minion_ai
      g.action_state = waitingAction
    } else {
      if g.haunts_ai.Active() {
        g.active_ai = g.haunts_ai
        g.action_state = waitingAction
      }
    }
  } else {
    if g.explorers_ai.Active() {
      g.active_ai = g.explorers_ai
      g.action_state = waitingAction
    }
  }
  if g.action_state == waitingAction {
    select {
    case exec := <-g.active_ai.ActionExecs():
      if exec != nil {
        g.current_exec = exec
      } else {
        g.action_state = noAction
        // TODO: indicate that the master ai can go now
      }

    default:
    }
  }
  if !g.player_active && g.action_state == noAction && !g.explorers_ai.Active() && !g.haunts_ai.Active() && !g.minion_ai.Active() {
    g.turn_state = turnStateMainPhaseOver
    base.Log().Printf("ScriptComm: change to turnStateMainPhaseOver")
    g.comm.game_to_script <- nil
    base.Log().Printf("ScriptComm: sent nil")
  }
}

func (g *Game) doLos(dist int, line [][2]int, los [][]bool) {
  var x0, y0, x, y int
  var room0, room *house.Room
  x, y = line[0][0], line[0][1]
  if x < 0 || y < 0 || x >= len(los) || y >= len(los[x]) {
    return
  }
  los[x][y] = true
  room = roomAt(g.House.Floors[0], x, y)
  for _, p := range line[1:] {
    x0, y0 = x, y
    x, y = p[0], p[1]
    if x < 0 || y < 0 || x >= len(los) || y >= len(los[x]) {
      return
    }
    room0 = room
    room = roomAt(g.House.Floors[0], x, y)
    if room == nil {
      return
    }
    if x == x0 || y == y0 {
      if room0 != nil && room0 != room && !connected(room, room0, x, y, x0, y0) {
        return
      }
    } else {
      roomA := roomAt(g.House.Floors[0], x0, y0)
      roomB := roomAt(g.House.Floors[0], x, y0)
      roomC := roomAt(g.House.Floors[0], x0, y)
      if roomA != nil && roomB != nil && roomA != roomB && !connected(roomA, roomB, x0, y0, x, y0) {
        return
      }
      if roomA != nil && roomC != nil && roomA != roomC && !connected(roomA, roomC, x0, y0, x0, y) {
        return
      }
      if roomB != nil && room != roomB && !connected(room, roomB, x, y, x, y0) {
        return
      }
      if roomC != nil && room != roomC && !connected(room, roomC, x, y, x0, y) {
        return
      }
    }
    furn := furnitureAt(room, x-room.X, y-room.Y)
    if furn != nil && furn.Blocks_los {
      return
    }
    dist -= 1 // or whatever
    if dist <= 0 {
      return
    }
    los[x][y] = true
  }
}

func (g *Game) TeamLos(x, y int) bool {
  if x < 0 || y < 0 || x >= len(g.los.merger) || y >= len(g.los.merger[x]) {
    return false
  }
  return g.los.merger[x][y]
}

func (g *Game) mergeLos(side Side) {
  var pix [][]byte
  switch side {
  case SideHaunt:
    pix = g.los.denizens.tex.Pix()
  case SideExplorers:
    pix = g.los.intruders.tex.Pix()
  default:
    base.Error().Printf("Unable to mergeLos on side %d.", side)
    return
  }
  for i := range g.los.full_merger {
    g.los.full_merger[i] = false
  }
  for _, ent := range g.Ents {
    if ent.Side() != side {
      continue
    }
    if ent.los == nil {
      continue
    }
    for i := ent.los.minx; i <= ent.los.maxx; i++ {
      for j := ent.los.miny; j <= ent.los.maxy; j++ {
        if ent.los.grid[i][j] {
          g.los.merger[i][j] = true
        }
      }
    }
  }
  for i := 0; i < len(pix); i++ {
    for j := 0; j < len(pix); j++ {
      if g.los.merger[i][j] {
        continue
      }
      if pix[i][j] >= house.LosVisibilityThreshold {
        pix[i][j] = house.LosVisibilityThreshold - 1
      }
    }
  }
  for i := range g.los.merger {
    for j := range g.los.merger[i] {
      if g.los.merger[i][j] {
        if pix[i][j] < house.LosVisibilityThreshold {
          pix[i][j] = house.LosVisibilityThreshold
        }
      }
    }
  }
}

// This is the function used to determine LoS.  Nothing else should try to
// make any attempts at doing so.  Eventually this should be replaced with
// something more sensible and faster, so everyone needs to use this so that
// everything stays in sync.
func (g *Game) DetermineLos(x, y, los_dist int, grid [][]bool) {
  for i := range grid {
    for j := range grid[i] {
      grid[i][j] = false
    }
  }
  minx := x - los_dist
  miny := y - los_dist
  maxx := x + los_dist
  maxy := y + los_dist
  line := make([][2]int, los_dist)
  for vx := minx; vx <= maxx; vx++ {
    line = line[0:0]
    bresenham(x, y, vx, miny, &line)
    g.doLos(los_dist, line, grid)
    line = line[0:0]
    bresenham(x, y, vx, maxy, &line)
    g.doLos(los_dist, line, grid)
  }
  for vy := miny; vy <= maxy; vy++ {
    line = line[0:0]
    bresenham(x, y, minx, vy, &line)
    g.doLos(los_dist, line, grid)
    line = line[0:0]
    bresenham(x, y, maxx, vy, &line)
    g.doLos(los_dist, line, grid)
  }
}

func (g *Game) UpdateEntLos(ent *Entity, force bool) {
  if ent.los == nil || ent.Stats == nil {
    return
  }
  ex, ey := ent.Pos()
  if !force && ex == ent.los.x && ey == ent.los.y {
    return
  }
  ent.los.x = ex
  ent.los.y = ey

  g.DetermineLos(ex, ey, ent.Stats.Sight(), ent.los.grid)

  ent.los.minx = len(ent.los.grid)
  ent.los.miny = len(ent.los.grid)
  ent.los.maxx = 0
  ent.los.maxy = 0
  for i := range ent.los.grid {
    for j := range ent.los.grid[i] {
      if ent.los.grid[i][j] {
        if i < ent.los.minx {
          ent.los.minx = i
        }
        if j < ent.los.miny {
          ent.los.miny = j
        }
        if i > ent.los.maxx {
          ent.los.maxx = i
        }
        if j > ent.los.maxy {
          ent.los.maxy = j
        }
      }
    }
  }
}

// Uses Bresenham's alogirthm to determine the points to rasterize a line from
// x,y to x2,y2.
func bresenham(x, y, x2, y2 int, res *[][2]int) {
  dx := x2 - x
  if dx < 0 {
    dx = -dx
  }
  dy := y2 - y
  if dy < 0 {
    dy = -dy
  }

  steep := dy > dx
  if steep {
    x, y = y, x
    x2, y2 = y2, x2
    dx, dy = dy, dx
  }

  err := dx >> 1
  cy := y

  xstep := 1
  if x2 < x {
    xstep = -1
  }
  ystep := 1
  if y2 < y {
    ystep = -1
  }
  for cx := x; cx != x2; cx += xstep {
    if !steep {
      *res = append(*res, [2]int{cx, cy})
    } else {
      *res = append(*res, [2]int{cy, cx})
    }
    err -= dy
    if err < 0 {
      cy += ystep
      err += dx
    }
  }
  if !steep {
    *res = append(*res, [2]int{x2, cy})
  } else {
    *res = append(*res, [2]int{cy, x2})
  }
}
