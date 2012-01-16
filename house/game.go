package house

import (
  "glop/gui"
)

type GamePanel struct {
  *gui.HorizontalTable

  house  *House
  viewer *HouseViewer
}

func MakeGamePanel() *GamePanel {
  var gp GamePanel
  gp.HorizontalTable = gui.MakeHorizontalTable()
  return &gp
}

func (gp *GamePanel) LoadHouse(name string) {
  gp.house = MakeHouse(name)
  if gp.viewer != nil {
    gp.HorizontalTable.RemoveChild(gp.viewer)
  }
  gp.viewer = MakeHouseViewer(gp.house, 62)
  gp.HorizontalTable.AddChild(gp.viewer)
}

func (gp *GamePanel) GetViewer() Viewer {
  return gp.viewer
}
