package house

import (
  "glop/util/algorithm"
  "glop/gui"
  "glop/gin"
)

type WallPanel struct {
  *gui.VerticalTable
  room *Room
  viewer *RoomViewer

  wall_texture *WallTexture
  prev_wall_texture *WallTexture
  drag_anchor struct{ X,Y float32 }
  drop_on_release bool
  select_mode selectMode
  selected_walls map[int]bool
}

func MakeWallPanel(room *Room, viewer *RoomViewer) *WallPanel {
  var wp WallPanel
  wp.room = room
  wp.viewer = viewer
  wp.VerticalTable = gui.MakeVerticalTable()
  wp.selected_walls = make(map[int]bool)

  fnames := GetAllWallTextureNames()
  for i := range fnames {
    name := fnames[i]
    wp.VerticalTable.AddChild(gui.MakeButton("standard", name, 300, 1, 1, 1, 1, func(t int64) {
      wt := MakeWallTexture(name)
      if wt == nil { return }
      wp.viewer.Temp.WallTexture = wt
      wp.viewer.Temp.WallTexture.X = 5
      wp.viewer.Temp.WallTexture.Y = 5
      wp.drag_anchor.X = 0
      wp.drag_anchor.Y = 0
      wp.drop_on_release = false
    }))
  }

  return &wp
}

func (w *WallPanel) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if w.VerticalTable.Respond(ui, group) {
    return true
  }
  if found,event := group.FindEvent(gin.Escape); found && event.Type == gin.Press {
    if w.viewer.Temp.WallTexture != nil {
      w.viewer.Temp.WallTexture = nil
    }
    return true
  }
  if found,event := group.FindEvent(gin.MouseLButton); found {
    if w.viewer.Temp.WallTexture != nil && (event.Type == gin.Press || (event.Type == gin.Release && w.drop_on_release)) {
      w.room.WallTextures = append(w.room.WallTextures, w.viewer.Temp.WallTexture)
      w.viewer.Temp.WallTexture = nil
    } else if w.viewer.Temp.WallTexture == nil && event.Type == gin.Press {
      x,y := w.viewer.WindowToBoard(event.Key.Cursor().Point())
      // w.viewer.Temp.WallTexture = w.viewer.SelectWallTextureAt(event.Key.Cursor().Point())
      w.viewer.Temp.WallTexture = nil
      if w.viewer.Temp.WallTexture != nil {
        w.prev_wall_texture = new(WallTexture)
        *w.prev_wall_texture = *w.viewer.Temp.WallTexture
      }
      w.room.WallTextures = algorithm.Choose(w.room.WallTextures, func(a interface{}) bool {
        return a.(*WallTexture) != w.viewer.Temp.WallTexture
      }).([]*WallTexture)
      w.drop_on_release = true
      if w.viewer.Temp.WallTexture != nil {
        w.drag_anchor.X = x - w.viewer.Temp.WallTexture.X
        w.drag_anchor.Y = y - w.viewer.Temp.WallTexture.Y
      }
    } else if event.Type == gin.Release {
      w.select_mode = modeNoSelect
    }
    return true
  }
  return false
}

func (w *WallPanel) Think(ui *gui.Gui, t int64) {
  if w.viewer.Temp.WallTexture != nil {
    mx,my := gin.In().GetCursor("Mouse").Point()
    x,y := w.viewer.WindowToBoard(mx, my)
    w.viewer.Temp.WallTexture.X = x - w.drag_anchor.X
    w.viewer.Temp.WallTexture.Y = y - w.drag_anchor.Y
  }
  w.VerticalTable.Think(ui, t)
}

func (w *WallPanel) Collapse() {
  if w.viewer.Temp.WallTexture != nil && w.prev_wall_texture != nil {
    w.room.WallTextures = append(w.room.WallTextures, w.prev_wall_texture)
  }
  w.prev_wall_texture = nil
  w.viewer.Temp.WallTexture = nil
  w.select_mode = modeNoSelect
}

func (w *WallPanel) Expand() {

}

