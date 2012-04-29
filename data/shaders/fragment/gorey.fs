uniform sampler2D tex;
varying vec3 pos;

void main() {
  // Only need to sample one color here - we're assuming the texture is black
  // and white
  float grey = texture2D(tex, gl_TexCoord[0].xy).r;

  // Noise based on three parameters
  // gl_TexCoord[0].s * 100.0: Gives some variance to avoid annoying artifacts
  // pos.x * 90.0: Prevents lines along the x axis
  // pos.y * 8.0: Stretches the lines along the y axis
  float n = noise1(vec3(gl_TexCoord[0].s * 100.0, pos.x * 90.0, pos.y * 8.0));
  n += grey;

  // Basically if something is light, we make it lighter, if it is dark, we
  // make it darker
  if (n > 0.5) {
    if (n > 0.9) {
      n = 1.0 - n;
      grey = 1.0 - n*n;
    } else {
      grey = n;
    }
  } else {
    if (n < 0.1) {
      grey = n*n;
    } else {
      grey = n;
    }
  }

  gl_FragColor = vec4(grey, grey, grey, 1.0);
}
