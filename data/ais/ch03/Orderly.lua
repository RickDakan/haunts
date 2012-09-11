function Think()
  while CrushIntruder(nil, nil, "Sedate", nil, nil) do
  end
  target = GetTarget()
  if target then
  	print("doingstuff")
    HeadTowards(target.Pos)
    print("Didstuff")
  end
  MoveLikeZombie()
  return false
end
