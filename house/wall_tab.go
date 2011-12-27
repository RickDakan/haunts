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
  drag_anchor struct{ pos,height float32 }
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
      wp.wall_texture = wt
      wp.wall_texture.Pos = 0.4
      wp.wall_texture.Height = 3
      wp.drag_anchor.pos = 0
      wp.drag_anchor.height = 0
      wp.drop_on_release = false
      wp.viewer.SetTempWallTexture(wt)
    }))
  }

  return &wp
}

func (w *WallPanel) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if w.VerticalTable.Respond(ui, group) {
    return true
  }
  if found,event := group.FindEvent(gin.Escape); found && event.Type == gin.Press {
    if w.wall_texture != nil {
      w.viewer.SetTempWallTexture(nil)
      w.wall_texture = nil
    }
    return true
  }
  if found,event := group.FindEvent(gin.MouseLButton); found {
    if w.wall_texture != nil && (event.Type == gin.Press || (event.Type == gin.Release && w.drop_on_release)) {
      w.viewer.SetTempWallTexture(nil)
      w.viewer.AddWallTexture(w.wall_texture)
      w.room.WallTextures = append(w.room.WallTextures, w.wall_texture)
      w.wall_texture = nil
    } else if w.wall_texture == nil && event.Type == gin.Press {
      pos,height := w.viewer.WindowToWall(event.Key.Cursor().Point())
      w.wall_texture = w.viewer.SelectWallTextureAt(event.Key.Cursor().Point())
      if w.wall_texture != nil {
        w.prev_wall_texture = new(WallTexture)
        *w.prev_wall_texture = *w.wall_texture
      } else {
        index := int(pos * float32(len(w.room.WallData)))
        if index >= 0 && index < len(w.room.WallData) {
          if w.selected_walls[index] {
            delete(w.selected_walls, index)
            w.select_mode = modeDeselect
          } else {
            w.selected_walls[index] = true
            w.select_mode = modeSelect
          }
        }
      }
      w.room.WallTextures = algorithm.Choose(w.room.WallTextures, func(a interface{}) bool {
        return a.(*WallTexture) != w.wall_texture
      }).([]*WallTexture)
      w.drop_on_release = true
      if w.wall_texture != nil {
        w.drag_anchor.pos = pos - w.wall_texture.Pos
        w.drag_anchor.height = height - float32(w.wall_texture.Height)
      }
    } else if event.Type == gin.Release {
      w.select_mode = modeNoSelect
    }
    return true
  }
  cursor := group.Events[0].Key.Cursor()
  if cursor != nil && cursor.Name() == "Mouse" {
    if w.select_mode == modeSelect || w.select_mode == modeDeselect {
      pos,_ := w.viewer.WindowToWall(cursor.Point())
      index := int(pos * float32(len(w.room.WallData)))
      if index >= 0 && index < len(w.room.WallData) {
        if w.select_mode == modeSelect {
          w.selected_walls[index] = true
          w.select_mode = modeSelect
        } else {
          delete(w.selected_walls, index)
          w.select_mode = modeDeselect
        }
      }
    }
  }
  return false
}

func (w *WallPanel) Think(ui *gui.Gui, t int64) {
  if w.wall_texture != nil {
    mx,my := gin.In().GetCursor("Mouse").Point()
    pos,height := w.viewer.WindowToWall(mx, my)
    w.wall_texture.Pos = pos - w.drag_anchor.pos
    w.wall_texture.Height = height - w.drag_anchor.height
  }
  w.VerticalTable.Think(ui, t)
}

func (w *WallPanel) Collapse() {
  if w.wall_texture != nil && w.prev_wall_texture != nil {
    w.viewer.AddWallTexture(w.prev_wall_texture)
  }
  w.viewer.SetTempWallTexture(nil)
  w.prev_wall_texture = nil
  w.wall_texture = nil
  w.select_mode = modeNoSelect
}

func (w *WallPanel) Expand() {

}

