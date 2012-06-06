denizens = activeNonMinions()
while denizens[1] do
  execNonMinion(denizens[1])
  denizens = activeNonMinions()
end
