function pursue()
  denizens = NearestNEntities (50, "denizen")
  for _, denizen in pairs (denizens) do
    target = Me().Info.LastEntityThatIAttacked
    if Exists(target) then
      return target
    end
  end
  return nil
end

function nearest()
  denizens = NearestNEntities (10, "denizen")
  for _, denizen in pairs (denizens) do
    return denizen
  end
  return nil
end

  
--target enemy that attacked nearest ally

function targetAllyAttacker()
  allies = NearestNEntities (10, "intruder")
  for _, ally in pairs (allies) do
    target = ally.Info.LastEntityThatAttackedMe
    if Exists(target) then
      return target
    end
  end 
  return nil
end

--target enemy your allies are already attacking
function targetAllyTarget()
  allies = NearestNEntities (10, "intruder")
  for _, ally in pairs (allies) do
    target = ally.Info.LastEntityThatIAttacked
    if Exists(target) then
      return target
    end
  end 
  return nil
end
  

-- target lowest stat
-- stat is looking for Corpus, Ego, HpCur, HpMax, ApCur, ApMax
function targetLowestStat(stat)
  denizens = NearestNEntities (50, "denizen")
  target = nil
  min = 10000
  for _, denizen in pairs (denizens) do
    if denizen[stat] < min then
      min = denizen[stat]
      target = denizen
    end
  end
  return target, min
end

--- Target Highest Stat
-- HOW?


function targetHighestStat(stat)
  denizens = NearestNEntities (50, "denizen")
  max = 0
  for _, denizen in pairs (denizens) do
    if denizen[stat] > max then
      max = denizen[stat]
      target = denizen
    end
  end
  return target, max
end



--target enemy with condition
-- has is a boolean that indicates whether you want the target to have the condition
-- true - the target will have the condition
-- false - the target will not have the condition

function targetHasCondition(has, condition)
  denizens = NearestNEntities (50, "denizen")
  for _, denizen in pairs (intruders) do
    if has and denizen[condition] then
      return denizen
    end
    if not has and not denizen[condition] then
      return denizen
    end
  end
end


-- BUFFing friends. Find friends who need a condition

function allyHasCondition(has, condition)
  allies = NearestNEntities(10, "intruder")
  for _, ally in pairs (allies) do
    if has and ally[condition] then
      return ally
    end
    if not has and not ally[condition] then
      return ally
    end
  end
end

function aoePlaceAndAttack(attack, spec)
  gz = BestAoeAttackPos (attack, Me().ApCur - Me().Actions[attack].Ap, spec)
  dsts = AllPathablePoints(Me().Pos, gz, 1, Me().Actions[attack].Range)
  DoMove(dsts, 1000)
  if RangedDistBetweenPositions (Me().Pos, gz) > Me().Actions[attack].Range then
    return
  else
    DoAoeAttack(attack, gz)
  end
end

