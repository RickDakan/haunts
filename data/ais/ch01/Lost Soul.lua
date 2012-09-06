function Think()
  while CrushIntruder(nil, nil, nil, "Ectoplasmic Discharge", nil) do
  end
  target = GetTarget()
  if target then
    ret = HeadTowards(target.Pos)
  end
  MoveLikeZombie()
  return false
end
