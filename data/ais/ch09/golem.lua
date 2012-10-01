function Think()
  target = GetDevice()
  if target then
  	print("doingstuff")
    HeadTowards(target)
    print("Didstuff")
  end
  return false
end
