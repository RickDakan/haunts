function Think()
  while CrushIntruder(nil, nil, "Chill Touch", nil, nil) do
  end
  target = GetDevice()
  if target then
  	print("doingstuff")
    HeadTowards(target)
    print("Didstuff")
  end
  return false
end
