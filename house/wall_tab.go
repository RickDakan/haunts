package house

import (
  "glop/util/algorithm"
  "glop/gui"
  "glop/gin"
)

type WallPanel struct {
  *gui.VerticalTable
  room *roomDef
  viewer *RoomViewer

  wall_texture *WallTexture
  prev_wall_texture *WallTexture
  drag_anchor struct{ X,Y float32 }
  selected_walls map[int]bool
}

func MakeWallPanel(room *roomDef, viewer *RoomViewer) *WallPanel {
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
    }))
  }

  return &wp
}

func (w *WallPanel) textureNear(wx,wy int) *WallTexture {
  for _,tex := range w.room.WallTextures {
    var xx,yy float32
    if tex.X > float32(w.room.Size.Dx) {
      xx,yy,_ = w.viewer.modelviewToRightWall(float32(wx), float32(wy))
    } else if tex.Y > float32(w.room.Size.Dy) {
      xx,yy,_ = w.viewer.modelviewToLeftWall(float32(wx), float32(wy))
    } else {
      xx,yy,_ = w.viewer.modelviewToBoard(float32(wx), float32(wy))
    }
    dx := float32(tex.Texture.Data().Dx) / 100 / 2
    dy := float32(tex.Texture.Data().Dy) / 100 / 2
    if xx > tex.X - dx && xx < tex.X + dx && yy > tex.Y - dy && yy < tex.Y + dy {
      return tex
    }
  }
  return nil
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
  if found,event := group.FindEvent(gin.MouseWheelVertical); found {
    if w.viewer.Temp.WallTexture != nil {
      w.viewer.Temp.WallTexture.Rot += float32(event.Key.CurPressAmt() / 100)
    }
  }
  if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    if w.viewer.Temp.WallTexture != nil {
      w.room.WallTextures = append(w.room.WallTextures, w.viewer.Temp.WallTexture)
      w.viewer.Temp.WallTexture = nil
    } else if w.viewer.Temp.WallTexture == nil {
      w.viewer.Temp.WallTexture = w.textureNear(event.Key.Cursor().Point())
      if w.viewer.Temp.WallTexture != nil {
        w.prev_wall_texture = new(WallTexture)
        *w.prev_wall_texture = *w.viewer.Temp.WallTexture
      }
      w.room.WallTextures = algorithm.Choose(w.room.WallTextures, func(a interface{}) bool {
        return a.(*WallTexture) != w.viewer.Temp.WallTexture
      }).([]*WallTexture)
      if w.viewer.Temp.WallTexture != nil {
        wx,wy := w.viewer.BoardToWindow(w.viewer.Temp.WallTexture.X, w.viewer.Temp.WallTexture.Y)
        px,py := event.Key.Cursor().Point()
        w.drag_anchor.X = float32(px) - wx - 0.5
        w.drag_anchor.Y = float32(py) - wy - 0.5
      }
    }
    return true
  }
  return false
}

func (w *WallPanel) Think(ui *gui.Gui, t int64) {
  if w.viewer.Temp.WallTexture != nil {
    px,py := gin.In().GetCursor("Mouse").Point()
    tx := float32(px) - w.drag_anchor.X
    ty := float32(py) - w.drag_anchor.Y
    bx,by := w.viewer.WindowToBoard(int(tx), int(ty))
    w.viewer.Temp.WallTexture.X = bx
    w.viewer.Temp.WallTexture.Y = by
  }
  w.VerticalTable.Think(ui, t)
}

func (w *WallPanel) Collapse() {
  if w.viewer.Temp.WallTexture != nil && w.prev_wall_texture != nil {
    w.room.WallTextures = append(w.room.WallTextures, w.prev_wall_texture)
  }
  w.prev_wall_texture = nil
  w.viewer.Temp.WallTexture = nil
}

func (w *WallPanel) Expand() {
}

func (w *WallPanel) Reload() {
}

