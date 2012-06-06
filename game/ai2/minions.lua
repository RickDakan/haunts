minions = activeMinions
for minions[1] != nil do
  execMinion(minions[1])
  minions = activeMinions
end


intruder = nearestNEntities(1, "intruder")
if 
if intruder[1] != nil then
  doBasicAttack("rawrcakes", intruder[1])
end
