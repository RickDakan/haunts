function Think()
  minions = activeMinions()
  print("fooo?")
  while minions[1] do
  	print("fooo2")
    execMinion(minions[1])
    minions = activeMinions()
  end
end
