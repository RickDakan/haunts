while true do
  intruders = nearestNEntities(3, "intruder")
  if intruders[1] then
    res = doBasicAttack("Chill Touch", intruders[1])
  end
end
