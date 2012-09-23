function Think()
  while CrushIntruder(nil, nil, "Ectoplasmic Discharge", nil, nil) do
  end
  target = GetTarget()
  if target then
    HeadTowards(target.Pos)
  end
  MoveLikeZombie()
  return false
end
