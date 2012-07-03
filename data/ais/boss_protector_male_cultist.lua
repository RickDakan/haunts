-- this function looks for the first intruder in intruders that is close
-- enough to master to hit it with some ranged basic attack that it has.
function protectMaster(master, intruders)
  for _, intruder in pairs(intruders) do
    -- Check for the longest range basic attack that the intruder has that
    -- does positive damage.
    range = 1
    actions = getActions(intruder)
    for action, _ in pairs(actions) do
      stats = getBasicAttackStats(intruder, action)
      if stats then
        if stats.range > range and stats.damage > 0 then
          range = stats.range
        end
      end
    end

    -- Check if the intruder is close enough to the master to hit it with
    -- a basic attack.
    if rangedDistBetweenEntities(master, intruder) <= range then
      -- We found an intruder that is too close to the master, so we will go
      -- after him.
      ps = allPathablePoints(Me.Pos, pos(intruder), 1, 1)
      if ps[1] then
        loc = doMove(ps, 1000)
        if loc then
          return intruder
        end
        -- If we failed to move then we will try for the next one, maybe there
        -- just wasn't a way to path to it right now.
      end
    end
  end

  -- Indicates that there was no one that the master needs protection from, or
  -- that we just aren't in a position to protect him right now.
  return nil
end

function Think()
  intruders = nearestNEntities(10, "intruder")
  master = nearestNEntities(1, "master")[1]

  -- If there are no intruders then we just stay put.
  if table.getn(intruders) == 0 then
    return
  end

  -- If there is a master in our los then we make sure to protect him
  if master then
    target = protectMaster(master, intruders)
    if target then
      while exists(target) do
        res = doBasicAttack("Kick", target)
        if res == nil then
          return
        end
      end
      -- We took out the target, so check again for a new target
      Think()
    end
  end

  -- If we made it here then we are free to just attack the nearest intruder
  intruder = intruders[1]
  ps = allPathablePoints(Me.Pos, pos(intruder), 1, 1)
  if ps[1] then
    loc = doMove(ps, 1000)
  end
  if rangedDistBetweenEntities(Me, intruder) == 1 then
    while exists(intruder) do
      res = doBasicAttack("Kick", intruder)
      if res == nil then
        return
      end
    end
  else
    return
  end

  Think()
end

