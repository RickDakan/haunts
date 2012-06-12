while true do
  intruders = nearestNEntities(3, "intruder")

  if intruders[1] then
    if moveToAndMeleeAttack("Kick", intruders[1]) == nil then
      break
    end
  else
    break
  end
end
