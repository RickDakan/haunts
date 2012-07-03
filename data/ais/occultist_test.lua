--occultist test
-- Intruder vs. Denizen Functions List??


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

function aoePlaceAndAttack(attack, spec)
	me_stats = getEntityStats(Me)
	attack_stats = getAoeAttackStats(Me, attack)
	gz = bestAoeAttackPos (attack, me_stats.apCur - attack_stats.ap, spec)
	dsts = allPathablePoints(Me.Pos, gz, 1, attack_stats.range)
	doMove(dsts, 1000)
	if rangedDistBetweenPositions (Me.Pos, gz) > attack_stats.range then
		return
	else
		doAoeAttack(attack, gz)
	end
end

-- Check for Ammo on First Aid

function Think()
	intruders = nearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		if getEntityStats(intruder).hpCur <6 and rangedDistBetweenEntites (Me, intruder) <5 then
			if getBasicAttackStats(Me, "aid").ammo >0 then
				moveWithinRangeAndAttack (1, "Aid", intruder)
			end
		end
	end
	target = pursue()
	if target == nil then
		target = targetAllyTarget()
	end
	if target == nil then
		target = targetHighestStat("hpCur")
	end
	if getEntityStat(target).corpus > 9 then
		MoveWithinRangeAndAttack(2, "Exorcise", target)
	else
		MoveWithinRangeAndAttack (2, "Dire Curse", target)
	end
end
	
--Targetting AOE Abjuration here, looking for at least two dudes together, unless Master present
























