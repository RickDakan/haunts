uniform float dx;
uniform float dy;

// 0 = !temporary && !invalid
// 1 = temporary && !invalid
// 2 = temporary && invalid
uniform int temp_invalid;

void main() {
  float bound = 0.15;
  float tx = abs(gl_TexCoord[0].x);
  if (tx < bound / dx || tx > 1.0 - bound / dx) {
    gl_FragColor = gl_Color * vec4(1.0, 1.0, 1.0, 1.0);
  } else {
    float ty = abs(gl_TexCoord[0].y);
    if (ty < bound / dy || ty > 1.0 - bound / dy) {
      gl_FragColor = gl_Color * vec4(1.0, 1.0, 1.0, 1.0);
    } else {
      if (temp_invalid == 0) {
        gl_FragColor = vec4(0.0, 0.0, 0.0, 0.3);
      } else if (temp_invalid == 1) {
        gl_FragColor = vec4(0.0, 0.0, 1.0, 0.3);
      } else if (temp_invalid == 2) {
        gl_FragColor = vec4(1.0, 0.0, 0.0, 0.3);
      } else {
        gl_FragColor = vec4(1.0, 1.0, 0.0, 0.3);
      }
    }
  }
}
