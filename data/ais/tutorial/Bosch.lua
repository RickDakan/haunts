function Think()
  while CrushIntruder(nil, nil, nil, "Dire Curse", nil) do
  end
  target = GetTarget()
  if target then
    HeadTowards(target.Pos)
  end
  MoveLikeZombie()
  return false
end
