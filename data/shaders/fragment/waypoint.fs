uniform float radius;
uniform float time;

const float orbit = 0.4;

void main() {
  float nub_size = 0.2 / radius;
  float dist = length(gl_TexCoord[0].st - vec2(0.5, 0.5));
  if (dist > orbit + nub_size || dist < orbit - nub_size) {
    gl_FragColor = gl_Color * vec4(0.0, 0.0, 0.0, 0.0);
    return;
  }
  float angle = atan(gl_TexCoord[0].s - 0.5, gl_TexCoord[0].t - 0.5);
  angle += time / radius;
  angle += 3.1415926535;
  angle = mod(angle, 2.0*3.1415926535);
  angle /= 2.0*3.1415926535;
  float arclength = angle * radius * 3.0;
  float width = nub_size * 2.0;
  float part = arclength - floor(arclength);
  float x;
  if (part < 0.5) {
    x = part;
  } else {
    x = 1.0 - part;
  }
  x /= dist;
  float shade = 1.0 - length(vec2(x, (dist - orbit)/nub_size));
  if (shade < 0.0) {
    shade = 0.0;
  }
  shade = shade * shade;
  gl_FragColor = vec4(1.0, 0.0, 0.0, shade);
}
