package house

import (
  "glop/gui"
  "github.com/arbaal/mathgl"
  "math"
  "haunts/base"
)

type HouseViewer struct {
  gui.Childless
  gui.EmbeddedWidget
  gui.BasicZone
  gui.NonFocuser
  gui.NonResponder
  gui.NonThinker

  house *houseDef

  zoom,angle,fx,fy float32
  floor,ifloor mathgl.Mat4

  Temp struct {
    Room *Room
  }
}

func MakeHouseViewer(house *houseDef, angle float32) *HouseViewer {
  var hv HouseViewer
  hv.EmbeddedWidget = &gui.BasicWidget{ CoreWidget: &hv }
  hv.Request_dims.Dx = 100
  hv.Request_dims.Dy = 100
  hv.Ex = true
  hv.Ey = true
  hv.house = house
  hv.angle = angle
  hv.Zoom(1)
  return &hv
}

func (hv *HouseViewer) modelviewToBoard(mx, my float32) (x,y,dist float32) {
  mz := d2p(hv.floor, mathgl.Vec3{mx, my, 0}, mathgl.Vec3{0,0,1})
  v := mathgl.Vec4{X: mx, Y: my, Z: mz, W: 1}
  v.Transform(&hv.ifloor)
  return v.X, v.Y, mz
}

func (hv *HouseViewer) WindowToBoard(wx, wy int) (float32, float32) {
  hv.floor, hv.ifloor, _, _, _, _ = makeRoomMats(&roomDef{}, hv.Render_region, hv.fx, hv.fy, hv.angle, hv.zoom)

  fx,fy,_ := hv.modelviewToBoard(float32(wx), float32(wy))
  return fx, fy
}

// Changes the current zoom from e^(zoom) to e^(zoom+dz)
func (hv *HouseViewer) Zoom(dz float64) {
  if dz == 0 {
    return
  }
  exp := math.Log(float64(hv.zoom)) + dz
  exp = float64(clamp(float32(exp), 2.5, 5.0))
  hv.zoom = float32(math.Exp(exp))
}

func (hv *HouseViewer) Drag(dx, dy float64) {
  hv.floor, hv.ifloor, _, _, _, _ = makeRoomMats(&roomDef{}, hv.Render_region, hv.fx, hv.fy, hv.angle, hv.zoom)

  v := mathgl.Vec4{X: hv.fx, Y: hv.fy, W: 1}
  v.Transform(&hv.floor)
  v.X += float32(dx)
  v.Y += float32(dy)

  v.Z = d2p(hv.floor, mathgl.Vec3{v.X, v.Y, 0}, mathgl.Vec3{0,0,1})
  v.Transform(&hv.ifloor)
  hv.fx, hv.fy = v.X, v.Y
}

func (hv *HouseViewer) String() string {
  return "house viewer"
}

func (hv *HouseViewer) Draw(region gui.Region) {
  region.PushClipPlanes()
  defer region.PopClipPlanes()
  hv.Render_region = region

  var rooms rectObjectArray
  for _,room := range hv.house.Floors[0].Rooms {
    rooms = append(rooms, room)
  }
  if hv.Temp.Room != nil {
    rooms = append(rooms, hv.Temp.Room)
  }
  rooms = rooms.Order()

  drawPrep()
  for i := len(rooms) - 1; i >= 0; i-- {
    room := rooms[i].(*Room)
    // TODO: Would be better to be able to just get the floor mats alone
    m_floor,_,m_left,_,m_right,_ := makeRoomMats(room.roomDef, region, hv.fx - float32(room.X), hv.fy - float32(room.Y), hv.angle, hv.zoom)

    var cstack base.ColorStack
    if room == hv.Temp.Room {
      cstack.Push(0.5, 0.5, 1, 0.75)
    } else {
      cstack.Push(1, 1, 1, 1)
    }
    drawWall(room.roomDef, m_floor, m_left, m_right, nil, cstack)
    drawFloor(room.roomDef, m_floor, nil, cstack)
    drawFurniture(m_floor, room.roomDef.Furniture, nil, cstack)
    // drawWall(room *roomDef, wall *texture.Data, left, right mathgl.Mat4, temp *WallTexture)
  }
}














