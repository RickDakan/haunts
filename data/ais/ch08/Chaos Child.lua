function Think()
  -- while CrushIntruder(nil, nil, "Sedate", nil, nil) do
  -- end
  target = WaypointPos()
  if target then
  	print("doingstuff")
    HeadTowards(target)
    print("Didstuff")
  end
  return false
end
