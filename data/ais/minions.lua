minions = activeMinions()
while minions[1] do
  execMinion(minions[1])
  minions = activeMinions()
end
