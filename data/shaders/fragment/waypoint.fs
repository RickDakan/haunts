uniform float radius;

const float outer = 0.4;
const float inner = 0.2;
const float middle = (outer + inner) / 2.0;

void main() {
  float dist = length(gl_TexCoord[0].st - vec2(0.5, 0.5));
  if (dist > outer || dist < inner) {
    gl_FragColor = vec4(0.0, 0.0, 0.0, 0.0);
    return;
  }
  float val;
  if (dist > middle) {
    val = 1.0 - (dist - middle) / (outer - middle);
  } else {
    val = (dist - inner) / (middle - inner);
  }
  gl_FragColor = gl_Color * vec4(val, val, val, val);
}
