package texture

import (
  "image"
  "image/draw"
  _ "image/jpeg"
  _ "image/png"
  "os"
  "runtime"
  "sync"
  "github.com/runningwild/glop/render"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/opengl/gl"
  "github.com/runningwild/opengl/glu"
)

type Object struct {
  Path base.Path

  // this path is the last one that was loaded, so that if we change Path then
  // we know to reload the texture.
  path base.Path
  data *Data
}
func (o *Object) Data() *Data {
  if o.data == nil || o.path != o.Path {
    o.data = LoadFromPath(string(o.Path))
    o.path = o.Path
  }
  return o.data
}

type Data struct {
  dx,dy int
  texture gl.Texture
}
func (d *Data) Dx() int {
  return d.dx
}
func (d *Data) Dy() int {
  return d.dy
}

var textureList uint
var textureListSync sync.Once
func setupTextureList() {
  textureListSync.Do(func() {
    render.Queue(func() {
      textureList = gl.GenLists(1)
      gl.NewList(textureList, gl.COMPILE)
      gl.Begin(gl.QUADS)
        gl.TexCoord2d(0, 0)
        gl.Vertex2i(0, 0)

        gl.TexCoord2d(0, -1)
        gl.Vertex2i(0, 1)

        gl.TexCoord2d(1, -1)
        gl.Vertex2i(1, 1)

        gl.TexCoord2d(1, 0)
        gl.Vertex2i(1, 0)
      gl.End()
      gl.EndList()
    })
  })
}
// Renders the texture on a quad at the texture's natural size.
func (d *Data) RenderNatural(x, y int) {
  if textureList != 0 {
    gl.PushMatrix()
    gl.Enable(gl.TEXTURE_2D)
    d.Bind()
    gl.Translated(float64(x), float64(y), 0)
    gl.Scaled(float64(d.dx), float64(d.dy), 1)
    gl.CallList(textureList)
    gl.PopMatrix()
  }
}

func (d *Data) Bind() {
  if d.texture == 0 {
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

func finalizeData(d *Data) {
  render.Queue(func() {
    d.texture.Delete()
  })
}

func (m *Manager) LoadFromPath(path string) *Data {
  setupTextureList()
  if data,ok := m.registry[path]; ok {
    return data
  }
  var data Data
  m.registry[path] = &data

  f,err := os.Open(path)
  if err != nil {
    base.Warn().Printf("Unable to load texture '%s': %v", path, err)
    return &data
  }
  config,_,err := image.DecodeConfig(f)
  f.Close()
  f,_ = os.Open(path)
  data.dx = config.Width
  data.dy = config.Height

  go func() {
    im,_,err := image.Decode(f)
    f.Close()
    if err != nil {
      base.Warn().Printf("Unable to decode texture '%s': %v", path, err)
      return
    }

    rgba := image.NewRGBA(image.Rect(0, 0, data.dx, data.dy))
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
      glu.Build2DMipmaps(gl.TEXTURE_2D, 4, data.dx, data.dy, gl.RGBA, rgba.Pix)
    })
    runtime.SetFinalizer(&data, finalizeData)
  } ()

  return &data
}
