function Init()
  map = selectMap()
  print(map)
  loadHouse(map)
  spawnDude("Angry Shade", 20, 20)
  spawnDude("Teen", 10, 20)
  spawnDude("Teen", 10, 20)
  showMainBar(true)
  ents = getAllEnts()
end

function OnRound(intruders)
  print("Intruders", intruders)
end