
function SupportAllies(buf, cond)
  -- First make sure I'm always buffed
  if not Me.Conditions[cond] then
    return Do.BasicAttack(buf, Me)
  end

  -- Now make sure my teammates are buffed if they are in trouble
  allies = Utils.NearestNEntities(3, "intruder")
  for _, ally in pairs(allies) do
    if ally.HpCur <= 5 and not ally.Conditions[cond] then
      return Do.BasicAttack(buf, ally)
    end
  end

  return false
end

function CrushEnemies(debuf, cond, melee, ranged)
  enemies = Utils.NearestNEntities(10, "denizen")
  if table.getn(enemies) == 0 then
    return false
  end
  max_dist = Me.Actions[ranged].Range
  lowest_hp = 10000
  lowest_ent = nil
  for i, enemy in pairs(enemies) do
    dist = Utils.RangedDistBetweenEntities(Me, enemy)
    if dist and dist <= max_dist and enemy.HpCur < lowest_hp then
      lowest_hp = enemy.HpCur
      lowest_ent = enemy
    end
  end
  if lowest_ent == nil then
    return false
  end
  target = lowest_ent
  dist = Utils.RangedDistBetweenEntities(Me, target)
  if cond and not target.Conditions[cond] and dist <= Me.Actions[debuf].Range then
    return Do.BasicAttack(debuf, target)
  end
  attack = ranged
  if dist == 1 then
    attack = melee
  end
  return Do.BasicAttack(attack, target)
end

function CheckForRelic()
  objects = Utils.NearestNEntities(3, "object")
  if table.getn(objects) == 0 then
    return false
  end

  object = nil
  for i,obj in pairs(objects) do
    if obj.State == "ready" then
      object = obj
      break
    end
  end
  if object == nil then
    return false
  end

  ps = Utils.AllPathablePoints(Me.Pos, object.Pos, 1, Me.Actions.Interact.Range)
  valid, pos = Do.Move(ps, 1000)
  if not valid then
    return false
  end
  return Do.InteractWithObject(object)
end

function Leader()
  intruders = Utils.NearestNEntities(3, "intruder")
  for _, ent in pairs(intruders) do
    if ent.Name == Me.Master.Leader then
      return ent
    end
  end
  return nil
end

function LeadOrFollow()
  print("Think: LeadOrFollow - ", Me.Name)
  if Me.Master.Leader == Me.Name then
    -- Don't go off leading somewhere if there is something nearby that needs
    -- to be dealt with
    ent = Utils.NearestNEntities(1, "denizen")[1]
    if ent then
      print("Think: Nearest denizen: ", ent.Name)
      dist = Utils.RangedDistBetweenEntities(Me, ent)
      print("Think: Dist - ", dist)
      if dist < 5 then
        return false
      end
    end

    print("Think: Leading")
    objects = Utils.NearestNEntities(3, "object")
    print("Think:", table.getn(objects), "objects.")
    if table.getn(objects) == 0 then
      return false
    end
    -- Should only ever be one at a time, so just take the first one
    object = objects[1]
    print("Think: Object at", object.Pos.X, object.Pos.Y)
    return HeadTowards(object.Pos)
  else
    return Follow(Leader())
  end
end

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
  print("Think: ", table.getn(unexplored), "unexplored rooms")
  if table.getn(unexplored) == 0 then
    return nil  -- No more rooms to explore
  end

  -- find the room whose center has the shortest manhattan distance to our
  -- target.
  dist = 0-1
  target_room = nil
  for _, room in pairs(unexplored) do
    print("Think: Getting distance")
    print("Think: ", target.X, target.Y)
    print("Think: ", room.Pos.X, room.Pos.Y)
    print("Think: ", room.Dims.Dx, room.Dims.Dy)
    c = getCenter(room)
    print("Think: ", c.X, c.Y)
    d = distance(target, getCenter(room))
    print("Think: Distance - ", d)
    if dist == 0-1 or d < dist then
      dist = d
      target_room = room
      print("Think: Room at ", room.Pos.X, room.Pos.Y, "is closer to", target.X, target.Y)
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

function HeadTowards(target)
  current = Utils.RoomContaining(Me)
  if posIsInRegion(target, current) then
    print("Think: Already in the room with our target")
    ps = Utils.AllPathablePoints(Me.Pos, target, 1, 1)
    print("Think: pses", table.getn(ps))
    valid, pos = Do.Move(ps, 1000)
    return valid and pos
  end

  target_room = FindUnexploredRoomNear(target)
  print("Think: Target room - ", target_room)
  if target_room == nil then
    return false
  end
  print("Think: Heading towards room at", target_room.Pos.X, target_room.Pos.Y)
  path = Utils.RoomPath(current, target_room)
  if table.getn(path) == 0 then
    return false  -- No room path to the unexplored room - shouldn't happen
  end
  target_room = path[1]

  doors = Utils.AllDoorsBetween(current, target_room)
  print("Think: Found", table.getn(doors), "doors")
  if table.getn(doors) == 0 then
    return false   -- No doors to the next room, also shouldn't happen
  end


  -- If the door is closed then go to it and open it.
  if not Utils.DoorIsOpen(doors[1]) then
    ps = Utils.DoorPositions(doors[1])
    print("Think: Moving to door")
    complete = Do.Move(ps, 1000)
    if not complete then
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

function Follow(leader)
  ps = Utils.AllPathablePoints(Me.Pos, leader.Pos, 1, 3)
  valid, pos = Do.Move(ps, 1000)
  return valid and pos
end


