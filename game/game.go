package game

import (
  "glop/gui"
  "glop/gin"
  "glop/util/algorithm"
  "haunts/base"
  "haunts/house"
)

type GamePanel struct {
  *gui.HorizontalTable

  house  *house.HouseDef
  viewer *house.HouseViewer

  // Keep track of this so we know how much time has passed between
  // calls to Think()
  last_think int64

  game *Game
}

func MakeGamePanel() *GamePanel {
  var gp GamePanel
  gp.house = house.MakeHouseDef()
  gp.viewer = house.MakeHouseViewer(gp.house, 62)
  gp.HorizontalTable = gui.MakeHorizontalTable()
  gp.HorizontalTable.AddChild(gp.viewer)
  return &gp
}
func (gp *GamePanel) Think(ui *gui.Gui, t int64) {
  if gp.last_think == 0 {
    gp.last_think = t
  }
  dt := t - gp.last_think
  gp.last_think = t

  if gp.game.action_state == doingAction {
    res := gp.game.current_action.Maintain(dt)
    switch res {
      case Complete:
        gp.game.current_action.Cancel()
        gp.game.current_action = nil
        gp.game.action_state = noAction

      case InProgress:
      case CheckForInterrupts:
    }
  }
  gp.game.Think(dt)
}

func (gp *GamePanel) Draw(region gui.Region) {
  gp.HorizontalTable.Draw(region)

  // Do heads-up stuff
  region.PushClipPlanes()
  defer region.PopClipPlanes()
  if gp.game.selected_ent != nil {
    gp.game.selected_ent.DrawReticle(gp.viewer, gp.game.selected_ent.Side == gp.game.Side, true)
  }
  if gp.game.hovered_ent != nil {
    gp.game.hovered_ent.DrawReticle(gp.viewer, gp.game.hovered_ent.Side == gp.game.Side, false)
  }
}

func (g *Game) setupRespond(ui *gui.Gui, group gui.EventGroup) bool {
  if group.Events[0].Key.Id() >= '1' && group.Events[0].Key.Id() <= '9' {
    if group.Events[0].Type == gin.Press {
      index := int(group.Events[0].Key.Id() - '1')
      names := base.GetAllNamesInRegistry("entities")
      ents := algorithm.Map(names, []*Entity{}, func(a interface{}) interface{} {
        return MakeEntity(a.(string))
      }).([]*Entity)
      ents = algorithm.Choose(ents, func(a interface{}) bool {
        return a.(*Entity).Side == g.Side
      }).([]*Entity)
      if index >= 0 && index < len(ents) {
        g.viewer.RemoveDrawable(g.new_ent)
        g.new_ent = ents[index]
        g.viewer.AddDrawable(g.new_ent)
      }
    }
  }
  if g.new_ent != nil {
    x,y := gin.In().GetCursor("Mouse").Point()
    bx,by := g.viewer.WindowToBoard(x, y)
    g.new_ent.X = float64(int(bx))
    g.new_ent.Y = float64(int(by))
    if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
      ix,iy := int(g.new_ent.X), int(g.new_ent.Y)
      r := roomAt(g.house.Floors[0], ix, iy)
      if r == nil { return true }
      f := furnitureAt(r, ix - r.X, iy - r.Y)
      if f != nil { return true }
      for _,e := range g.Ents {
        x,y := e.Pos()
        if x == ix && y == iy { return true }
      }
      g.Ents = append(g.Ents, g.new_ent)
      g.new_ent = nil
    }
  }
  return false
}

func (gp *GamePanel) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if gp.HorizontalTable.Respond(ui, group) {
    return true
  }

  if found,event := group.FindEvent(base.GetDefaultKeyMap()["finish round"].Id()); found && event.Type == gin.Press {
    gp.game.OnRound()
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
        gp.game.hovered_ent = gp.game.Ents[i]
      }
    }
  }

  if found,_ := group.FindEvent(gin.Escape); found {
    if gp.game.selected_ent != nil {
      switch gp.game.action_state {
      case noAction:
        gp.game.selected_ent = nil
        return true

      case preppingAction:
        gp.game.action_state = noAction
        gp.game.current_action.Cancel()
        gp.game.current_action = nil
        return true

      case doingAction:
        // Do nothing - we don't cancel an action that's in progress
      }
    }
  }

  if gp.game.action_state == noAction {
    if found,_ := group.FindEvent(gin.MouseLButton); found {
      if gp.game.hovered_ent != nil && gp.game.hovered_ent.Side == gp.game.Side {
        gp.game.selected_ent = gp.game.hovered_ent
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
          if gp.game.current_action != nil {
            gp.game.current_action.Cancel()
          }
          gp.game.current_action = action
          gp.game.action_state = preppingAction
        }
      }
    }
  }

  return false
}

func (gp *GamePanel) LoadHouse(name string) {
  gp.HorizontalTable = gui.MakeHorizontalTable()
  gp.house = house.MakeHouse(name)
  if len(gp.house.Floors) == 0 {
    gp.house = house.MakeHouseDef()
  }
  gp.viewer = house.MakeHouseViewer(gp.house, 62)
  gp.game = makeGame(gp.house, gp.viewer)
  gp.viewer.Los_tex = gp.game.los_tex
  gp.HorizontalTable = gui.MakeHorizontalTable()
  gp.HorizontalTable.AddChild(gp.viewer)
}

func (gp *GamePanel) GetViewer() house.Viewer {
  return gp.viewer
}
