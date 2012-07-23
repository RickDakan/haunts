package base

import (
  "crypto/rand"
)

func MakeCMWC() *CMWC {
  var c CMWC
  c.Q = make([]uint64, 4096)
  c.C = 362436
  c.A = 18782
  return &c
}

type CMWC struct {
  Q []uint64
  A uint64
  C uint32
}

func (c *CMWC) Seed(seed int64) {
  for i := range c.Q {
    c.Q[i] = uint64(seed)
  }
}

func (c *CMWC) SeedWithDevRand() {
  buf := make([]byte, len(c.Q)*8)
  rand.Reader.Read(buf)
  for i := range c.Q {
    c.Q[i] = 0
    for j := 0; j < 8; j++ {
      c.Q[i] |= uint64(buf[i*8+j]) << uint(8*j)
    }
  }
}

func (c *CMWC) Int63() int64 {
  var t uint64
  var r uint32 = 0xfffffffe
  i := uint32(len(c.Q) - 1)
  i = (i + 1) & (uint32(len(c.Q)) - 1)
  t = c.A*c.Q[i] + uint64(c.C)
  c.C = uint32(t >> 32)
  x := t + uint64(c.C)
  if x < uint64(c.C) {
    x++
    c.C++
  }
  c.Q[i] = uint64(r) - x
  return int64(c.Q[i] >> 1)
}
