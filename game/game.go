package game

import (
  "glop/gui"
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
        gp.game.current_action = nil
        gp.game.action_state = noAction

      case InProgress:
      case CheckForInterrupts:
    }
  }
  gp.game.Think(dt)
}
func (gp *GamePanel) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if gp.HorizontalTable.Respond(ui, group) {
    return true
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

  if gp.game.action_state == noAction || gp.game.action_state == preppingAction {
    if len(group.Events) == 1 && group.Events[0].Key.Id() >= '1' && group.Events[0].Key.Id() <= '9' {
      index := int('1' - group.Events[0].Key.Id())
      if index >= 0 && index < len(gp.game.Ents[0].Actions) {
        action := gp.game.Ents[0].Actions[index]
        if action != gp.game.current_action && action.Prep(gp.game.Ents[0]) {
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
  gp.HorizontalTable.RemoveChild(gp.viewer)
  gp.house = house.MakeHouse(name)
  if len(gp.house.Floors) == 0 {
    gp.house = house.MakeHouseDef()
  }
  gp.viewer = house.MakeHouseViewer(gp.house, 62)
  gp.game = makeGame(gp.house, gp.viewer)
  gp.viewer.Los_tex = gp.game.los_tex
  gp.HorizontalTable.AddChild(gp.viewer)
}

func (gp *GamePanel) GetViewer() house.Viewer {
  return gp.viewer
}
