package texture

import (
  "gl"
  "gl/glu"
  "os"
  "image"
  _ "image/png"
  _ "image/jpeg"
  "image/draw"
  "glop/render"
)

type Data struct {
  Dx,Dy int
  texture gl.Texture

  // If there was an error loading this texture it will be stored here
  Err error
}

func (d *Data) Bind() {
  if d.Err != nil {
    println("ERror: ", d.Err.Error())
  }
  d.texture.Bind(gl.TEXTURE_2D)
}

var (
  manager Manager
)

func init() {
  manager.registry = make(map[string]*Data)
}

type Manager struct {
  registry map[string]*Data
}

func Reload() {
  manager.Reload()
}
func (m *Manager) Reload() {
}

func LoadFromPath(path string) *Data {
  return manager.LoadFromPath(path)
}
func (m *Manager) LoadFromPath(path string) *Data {
  if data,ok := m.registry[path]; ok {
    return data
  }
  var data Data
  m.registry[path] = &data

  go func() {
    f,err := os.Open(path)
    if err != nil {
      data.Err = err
      return
    }
    im,_,err := image.Decode(f)
    f.Close()
    if err != nil {
      data.Err = err
      return
    }

    data.Dx = im.Bounds().Dx()
    data.Dy = im.Bounds().Dy()
    rgba := image.NewRGBA(image.Rect(0, 0, data.Dx, data.Dy))
    draw.Draw(rgba, im.Bounds(), im, image.Point{0, 0}, draw.Over)
    render.Queue(func() {
      gl.Enable(gl.TEXTURE_2D)
      data.texture = gl.GenTexture()
      data.texture.Bind(gl.TEXTURE_2D)
      gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
      gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
      gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
      gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
      gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
      glu.Build2DMipmaps(gl.TEXTURE_2D, 4, data.Dx, data.Dy, gl.RGBA, rgba.Pix)
    })
  } ()

  return &data
}
