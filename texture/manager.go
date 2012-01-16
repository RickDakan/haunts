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

type Object struct {
  Path string `registry:"path"`

  // this path is the last one that was loaded, so that if we change Path then
  // we know to reload the texture.
  path string
  data *Data
}
func (o *Object) Data() *Data {
  if o.data == nil || o.path != o.Path {
    o.data = LoadFromPath(o.Path)
    o.path = o.Path
  }
  return o.data
}

type Data struct {
  Dx,Dy int
  texture gl.Texture

  // If there was an error loading this texture it will be stored here
  Err error
}

func (d *Data) Bind() {
  if d.Err != nil {
    if error_texture == 0 {
      gl.Enable(gl.TEXTURE_2D)
      error_texture = gl.GenTexture()
      error_texture.Bind(gl.TEXTURE_2D)
      gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
      gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
      gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
      gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
      gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
      pink := []byte{ 255, 0, 255, 255 }
      glu.Build2DMipmaps(gl.TEXTURE_2D, 4, 1, 1, gl.RGBA, pink)
    }
    error_texture.Bind(gl.TEXTURE_2D)
  } else {
    d.texture.Bind(gl.TEXTURE_2D)
  }
}

var (
  manager Manager
  error_texture gl.Texture
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
      println("Error loading ", path, " ", data.Err.Error())
      return
    }
    im,_,err := image.Decode(f)
    f.Close()
    if err != nil {
      data.Err = err
      println("Error loading ", path, " ", data.Err.Error())
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
