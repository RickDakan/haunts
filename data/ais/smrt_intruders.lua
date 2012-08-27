function Think()
  intruders = activeIntruders()
  while intruders[1] do
    execIntruder(intruders[1])
    intruders = activeIntruders()
  end
end
