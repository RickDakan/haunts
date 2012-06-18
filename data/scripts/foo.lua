function Init()
  map = selectMap()
  loadHouse(map)
  spawnDude("Angry Shade", 20, 20)
  spawnDude("Teen", 10, 20)
  dudes = placeDude("Foo-.*")
  for i, dude in pairs(dudes) do
    print(i, dude)
  end
  showMainBar(true)
  ents = getAllEnts()
end

function OnRound(intruders)
  print("Intruders", intruders)
end