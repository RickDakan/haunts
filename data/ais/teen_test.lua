--Teen
-- add in use Telepathic Shroud when near Goal
--UTIL Entity Folders - to alls
-- also WILL have Minions, Denizens, Intruders added to 

function pursue()
	denizens = nearestNEntities (50, "denizen")
	for _, denizen in pairs (denizens) do
		target = entityInfo(Me).lastEntityIAttacked
		if exists(target) then
			return target
		end
	end
	return nil
end

function nearest()
	denizens = nearestNEntities (10, "denizen")
	for _, denizen in pairs (denizens) do
		return denizen
	end
	return nil
end

	
--target enemy that attacked nearest ally

function targetAllyAttacker()
	allies = nearestNEntities (10, "intruder")
	for _, ally in pairs (allies) do
	  target = entityInfo(ally).lastEntityThatAttackedMe
	  if exists(target) then
	  	return target
	  end
	end	
	return nil
end

--target enemy your allies are already attacking
function targetAllyTarget()
	allies = nearestNEntities (10, "intruder")
	for _, ally in pairs (allies) do
	  target = entityInfo(ally).lastEntityIAttacked
	  if exists(target) then
	  	return target
	  end
	end	
	return nil
end
	

-- target lowest stat
-- stat is looking for corpus, ego, hpCur, hpMax, apCur, apMax
function targetLowestStat(stat)
	denizens = nearestNEntities (50, "denizen")
	target = nil
	min = 10000
	for _, denizen in pairs (denizens) do
		if getEntityStats(denizen) [stat] < min then
			min = getEntityStats(denizen) [stat]
			target = denizen
		end
	end
	return target, min
end

--- Target Highest Stat
-- HOW?


function targetHighestStat(stat)
	denizens = nearestNEntities (50, "denizen")
	max = 0
	for _, denizen in pairs (denizens) do
		if getEntityStats(denizen) [stat] > max then
			max = getEntityStats(denizen) [stat]
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
	denizens = nearestNEntities (50, "denizen")
	for _, denizen in pairs (intruders) do
		if has and getConditions(denizen) [condition] then
			return denizen
		end
		if not has and not getConditions(denizen) [condition] then
			return denizen
		end
	end
end


-- BUFFing friends. Find friends who need a condition

function allyHasCondition(has, condition)
	allies = nearestNEntities(10, "intruder")
	for _, ally in pairs (allies) do
		if has and getConditions(ally) [condition] then
			return ally
		end
		if not has and not getConditions(ally) [condition] then
			return ally
		end
	end
end





function Think()
	if getEntityStats(Me).hpCur < 5 and not getCondition(Me) ["Psychic Shroud"] then
		doBasicAttack ("Psychic Shroud", Me)
	end
	denizens = nearestNEntities (50, "denizen")
	for _, denizen in pairs (denizens) do
		if getEntityStats(denizen).corpus >10 and not getConditions(denizen) ["Telepathic Target"] then
			MoveWithinRangeAndAttack (1, "Telepathic Coordination", denizen)
		end
	end
	target = Pursue()
	if target == nil then
		target = retaliate()
	end
	if target == nil then
		target = targetAllyTarget()
	end
	if target == nil then
		target = targetLowestStat("hpCur")
	end
	if target == nil then
		return
	end
	if rangedDistBetweenEntities (Me, target) <2 then
		MoveWithinRangeAndAttack(1, "Kick", target)
	else
		MoveWithinRangeAndAttack (2, "Pistol", target)
	end
end

			
		