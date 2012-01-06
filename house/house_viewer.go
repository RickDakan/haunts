package house

import (
  "glop/gui"
  "github.com/arbaal/mathgl"
)

type HouseViewer struct {
  gui.Childless
  gui.EmbeddedWidget
  gui.BasicZone
  gui.NonFocuser
  gui.NonResponder
  gui.NonThinker

  house *houseDef

  mat,imat mathgl.Mat4
}

func MakeHouseViewer(house *houseDef, angle float32) *HouseViewer {
  var hv HouseViewer
  hv.EmbeddedWidget = &gui.BasicWidget{ CoreWidget: &hv }
  return &hv
}

func (hv *HouseViewer) Zoom(float64) {
  
}

func (hv *HouseViewer) Drag(float64,float64) {
  
}

func (hv *HouseViewer) String() string {
  return "house viewer"
}

func (hv *HouseViewer) Draw(region gui.Region) {
  // room := hv.house.Floors[0].Rooms[0].roomDef

  // TODO: Would be better to be able to just get the floor mats alone
  // m_floor,_,_,_,_,_ := makeRoomMats(room, region, 0, 0, 62, 1)
  
  // drawFloor(room, floor *texture.Data, nil)

  // drawWall(room *roomDef, wall *texture.Data, left, right mathgl.Mat4, temp *WallTexture)
}














