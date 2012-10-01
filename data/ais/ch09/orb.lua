function Think()
  target = GetTarget()
  if target then
  	print("doingstuff")
    HeadTowards(target)
    print("Didstuff")
  end
  return false
end
