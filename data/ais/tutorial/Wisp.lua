function Think()
  while CrushIntruder(nil, nil, nil, "Ectoplasmic Discharge", nil) do
  end
  target = GetTarget()
  if target then
    HeadTowards(target.Pos)
  end
  return false
end
