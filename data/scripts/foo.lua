function Init()
  map = selectMap()
  loadHouse(map)
  spawnDude("Angry Shade", 20, 20)
  spawnDude("Teen", 10, 20)
  placeDude()
  showMainBar(true)
  ents = getAllEnts()
end

function OnRound(intruders)
  print("Intruders", intruders)
end