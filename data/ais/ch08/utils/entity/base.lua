function getCenter(thing)
  x = thing.Pos.X + thing.Dims.Dx / 2
  y = thing.Pos.Y + thing.Dims.Dy / 2
  return {X=x, Y=y}
end

function distance(a, b)
  v1 = a.X - b.X
  if v1 < 0 then
    v1 = 0-v1
  end
  v2 = a.Y - b.Y
  if v2 < 0 then
    v2 = 0-v2
  end
  return v1 + v2
end

function FindUnexploredRoomNear(target)
  unexplored = Utils.NearbyUnexploredRooms()
  if table.getn(unexplored) == 0 then
    return nil  -- No more rooms to explore
  end

  -- find the room whose center has the shortest distance to our target.
  dist = 0-1
  target_room = nil
  for _, room in pairs(unexplored) do
    c = getCenter(room)
    d = distance(target, getCenter(room))
    if dist == 0-1 or d < dist then
      dist = d
      target_room = room
    end
  end

  return target_room
end

-- pos.X and pos.Y are the position in question
-- region.Pos.X, region.Pos.Y, region.Dims.Dx and region.Dims.Dy define
-- the region that we will check for pos in.
function posIsInRegion(pos, region)
  if pos.X < region.Pos.X then
    return false
  end
  if pos.Y < region.Pos.Y then
    return false
  end
  if pos.X >= region.Pos.X + region.Dims.Dx then
    return false
  end
  if pos.Y >= region.Pos.Y + region.Dims.Dy then
    return false
  end
  return true
end

-- Moves Me towards target, wherever target happens to be.
function HeadTowards(target)
  print("SCRIPT: HEadingTowards 1")
  current = Utils.RoomContaining(Me)
  
  if posIsInRegion(target, current) then
    print("SCRIPT: HEadingTowards - already in room!")
    ps = Utils.AllPathablePoints(Me.Pos, target, 1, 1)
    valid, pos = Do.Move(ps, 1000)
    return valid and pos
  end
  print("SCRIPT: HEadingTowards 2")

  target_room = FindUnexploredRoomNear(target)
  print("SCRIPT: HEADING found unexplored")
  if target_room == nil then
    return false
  end
  print("SCRIPT: HEadingTowards 3")
  path = Utils.RoomPath(current, target_room)
  if table.getn(path) == 0 then
    return false  -- No room path to the unexplored room - shouldn't happen
  end
  print("SCRIPT: HEadingTowards 4")
  target_room = path[1]

  doors = Utils.AllDoorsBetween(current, target_room)
  if table.getn(doors) == 0 then
    return false   -- No doors to the next room, also shouldn't happen
  end
  print("SCRIPT: HEadingTowards 5")


  -- If the door is closed then go to it and open it.
  if not Utils.DoorIsOpen(doors[1]) then
    ps = Utils.DoorPositions(doors[1])
    for i,p in pairs(ps) do
    end
    complete = Do.Move(ps, 1000)
    if not complete then
      return false
    end

    print("SCRIPT: ME.Sie.In", Me.Side.Intruder)
    if Me.Side.Intruder then
      -- Only open it if we have at least half of our max ap and our peeps
      -- are nearby
      if Me.ApCur < Me.ApMax / 2 then
        return false
      end
      intruders = Utils.NearestNEntities(3, "intruder")
      if table.getn(intruders) > 0 then
        if Utils.RangedDistBetweenEntities(Me, intruders[table.getn(intruders)]) > 5 then
          return false
        end
      end
    end

    if not Me.Actions["Interact"] then
      print("WISP: CANT OPEN DOORS")
      return false
    end
    if not Do.DoorToggle(doors[1]) then
      return false
    end
  end

  -- Now that we know the door is open, step into the next room.
  ps = Utils.RoomPositions(target_room)
  complete, other = Do.Move(ps, 1000)
  return complete
end

