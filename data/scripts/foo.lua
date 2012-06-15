function Init()
  loadHouse("milestone-manor.house")
  spawnDude("Angry Shade", 20, 20)
  spawnDude("Teen", 10, 20)
  spawnDude("Teen", 10, 20)
  showMainBar(true)
  ents = getAllEnts()
end

function OnRound(intruders)
  print("Intruders", intruders)
end