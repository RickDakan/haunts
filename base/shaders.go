package base

import (
  "sync"
  "unsafe"
  "github.com/runningwild/glop/render"
  gl "github.com/chsc/gogl/gl21"
)

const vertex_shader = `
  void main() {
    gl_Position = ftransform();
    gl_ClipVertex = gl_ModelViewMatrix * gl_Vertex;
    gl_FrontColor = gl_Color;
    gl_TexCoord[0].st = gl_MultiTexCoord0.st;
    gl_TexCoord[1].st = gl_MultiTexCoord1.st;
  }
`
const fragment_shader = `
  uniform sampler2D tex1;
  uniform sampler2D tex2;
  void main() {
    vec4 value1 = texture2D(tex1, gl_TexCoord[0].st);
    vec4 value2 = texture2D(tex2, gl_TexCoord[1].st);
    gl_FragColor = value1 * vec4(value2.w, value2.w, value2.w, gl_Color.w);
  }
`

var vertex_shader_object uint32
var fragment_shader_object uint32
var program_object uint32

func EnableShader(enable bool) {
  if enable {
    gl.UseProgram(program_object)
  } else {
    gl.UseProgram(0)
  }
}

var shaders_once sync.Once
func InitShaders() {
  shaders_once.Do(func() {
    render.Queue(func() {
      vertex_shader_object = gl.CreateShader(gl.VERTEX_SHADER)
      var prog []gl.Char
      for i := range vertex_shader {
        prog = append(prog, gl.Char(vertex_shader[i]))
      }
      pointer := &prog[0]
      length := int32(len(vertex_shader))
      gl.ShaderSource(vertex_shader_object, 1, (**gl.Char)(unsafe.Pointer(&pointer)), &length)
      gl.CompileShader(vertex_shader_object)
      var param int32
      gl.GetShaderiv(vertex_shader_object, gl.COMPILE_STATUS, &param)
      if param == 0 {
        Error().Printf("Failed to compile vertex shader")
        return
      }

      fragment_shader_object = gl.CreateShader(gl.FRAGMENT_SHADER)
      prog = prog[0:0]
      for i := range fragment_shader {
        prog = append(prog, gl.Char(fragment_shader[i]))
      }
      pointer = &prog[0]
      length = int32(len(fragment_shader))
      gl.ShaderSource(fragment_shader_object, 1, (**gl.Char)(unsafe.Pointer(&pointer)), &length)
      gl.CompileShader(fragment_shader_object)
      gl.GetShaderiv(fragment_shader_object, gl.COMPILE_STATUS, &param)
      if param == 0 {
        Error().Printf("Failed to compile fragment shader")
        return
      }

      // shader successfully compiled - now link
      program_object = gl.CreateProgram()
      gl.AttachShader(program_object, vertex_shader_object)
      gl.AttachShader(program_object, fragment_shader_object)
      gl.LinkProgram(program_object)
      gl.GetProgramiv(program_object, gl.LINK_STATUS, &param)
      if param == 0 {
        Error().Printf("Failed to link shader")
        return
      }

      gl.UseProgram(program_object)
      name := []byte("tex2\x00")
      tex2_loc := gl.GetUniformLocation(program_object, (*gl.Char)(unsafe.Pointer(&name[0])))
      gl.Uniform1i(tex2_loc, 1)
      gl.UseProgram(0)
    })
  })
}
