function Think()
  while CrushIntruder(nil, nil, "Chill Touch", nil, nil) do
  end
  print("SCRIPT: 1")
  target = GetTarget()
  print("SCRIPT: 2")
  if target then
  print("SCRIPT: Heading towards ", target.Pos.X, target.Pos.Y)
    HeadTowards(target.Pos)
  print("SCRIPT: 4")
  end
  print("SCRIPT: 5")
  MoveLikeZombie()
  return false
end
