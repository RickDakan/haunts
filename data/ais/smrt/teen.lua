
function Think()
  -- objects = Utils.NearestNEntities(1, "object")
  -- if table.getn(objects) > 0 then
  --   ps = Utils.AllPathablePoints(Me.Pos, objects[1].Pos, 1, 1)
  --   Do.Move(ps, 1000)
  --   return
  -- end
  unexplored = Utils.NearbyUnexploredRoom()
  if not unexplored then
    print("Couldn't find an unexplored room!")
    return
  end
  current = Utils.RoomContaining(Me)
  for k,v in pairs(unexplored) do
    print("Unexplored", k, v)
  end
  for k,v in pairs(current) do
    print("Current", k, v)
  end
  path = Utils.RoomPath(current, unexplored)
  if table.getn(path) == 0 then
    print("Unable to find a path to the unexplored room!")
    return
  end
  target = path[1]
  doors = Utils.AllDoorsBetween(current, target)
  if table.getn(doors) == 0 then
    print("Unable to find a door to the next room!")
    return
  end
  ps = Utils.DoorPositions(doors[1])
  Do.Move(ps, 1000)
  return
end
