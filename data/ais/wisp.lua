-- Depending on our ap and the cost of our attack we might be able to manage
-- more than one attack, so we will just loop this infinitely.
while true do
  -- We'll only concern ourselves with trying to attack the closest intruder,
  -- a more complicated script might want to keep track of more than one.
  intruders = nearestNEntities(1, "intruder")

  -- If there are no intruders in sight we just hang out and wait
  if not intruders[1] then
    break
  end

  mypos = Me.Pos
  hispos = intruders[1].Pos
  attack = "Ectoplasmic Discharge"
  stats = Me.Actions[attack]

  -- We want to be withing range to hit our target, but we don't want to be
  -- much closer than we need to be.  So if our range is 7 we will try to get
  -- within 5-7 of our target, if it is 4 we will get within 2-4, etc...
  min = stats.Range - 2
  if min < 1 then
    min = 1
  end

  -- This gives us all points that we could walk to right now that are within
  -- the appropriate ranged distance we are looking for (min and stats.Range)
  ps = AllPathablePoints(mypos, hispos, min, stats.Range)

  -- If there is no way we can path there then we give up.  Alternatively we
  -- could have kept track of other nearby intruders and tried to target one
  -- of them instead.
  if not ps[1] then
    break
  end

  -- We move to one of the target spaces, the closest space might be the one
  -- we are currently standing on, in which case we won't move and we will
  -- skip to attacking.
  doMove(ps, 1000)

  -- This does an attack with the basic attack we specified earlier on our
  -- target.  The result will have a value (res.hit) that is a boolean
  -- indicating whether or not our attack hit its target.
  res = doBasicAttack(attack, intruders[1])
  if res == nil then
    break
  end
end
