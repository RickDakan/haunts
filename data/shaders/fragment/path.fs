uniform sampler2D tex1;
uniform float threshold;
uniform float size;

void main() {
  vec2 t = size * gl_TexCoord[0].st;
  vec2 s = t;
  s = floor(s);
  s += vec2(0.5, 0.5);
  float dist = length(t-s);
  if (dist > 0.5) {
    gl_FragColor = vec4(0.0, 0.0, 0.0, 0.0);
    return;
  }

  float scale = 1.0;
  if (dist > 0.3) {
    float f = dist - 0.3;
    f /= 0.5 - 0.3;
    scale = 1.0 - f;
  }

  s /= size;
  vec4 path = texture2D(tex1, s);
  if (path.a == 0.0) {
    gl_FragColor = vec4(0.0, 0.0, 0.0, 0.0);
  } else if (path.a <= threshold) {
    gl_FragColor = vec4(1.0, 0.0, 0.0, gl_Color.a * scale);
  } else {
    gl_FragColor = vec4(0.3, 0.0, 0.0, gl_Color.a * scale);
  }
}
