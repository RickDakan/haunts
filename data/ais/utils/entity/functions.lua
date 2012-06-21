function moveWithinRangeAndAttack (min_range, attack, target)
	max_range = getBasicAttackStats(me(), attack).range
	if min_range > max_range then
		min_range = max_range
	end
	ps = allPathablePoints (pos(me()), pos(target), min_range, max_range)
	res = doMove (ps, 1000)
	if exists(target) then
		doBasicAttack(attack, target)
	end
end

function pursue()
	intruders = nearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		target = entityInfo(me()).lastEntityIAttacked
		if exists(target) then
			return target
		end
	end
	return nil
end

function nearest()
	intruders = nearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		return intruder
	end
	return nil
end


function retaliate()
	target = entityInfo(me()).lastEntityThatAttackedMe
	if exists(target) then
		return target
	end
	return nil
end

		
--target enemy that attacked nearest ally

function targetAllyAttacker()
	allies = nearestNEntities (50, "denizen")
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
	allies = nearestNEntities (50, "denizen")
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
	intruders = nearestNEntities (10, "intruder")
	target = nil
	min = 10000
	for _, intruder in pairs (intruders) do
		if getEntityStats(intruder) [stat] < min then
			min = getEntityStats(intruder) [stat]
			target = intruder
		end
	end
	return target, min
end

--- Target Highest Stat
-- HOW?

function targetHighestStat(stat)
	intruders = nearestNEntities (10, "intruder")
	max = 0
	for _, intruder in pairs (intruders) do
		if getEntityStats(intruder) [stat] > max then
			max = getEntityStats(intruder) [stat]
			target = intruder
		end
	end
	return target, max
end



--target enemy with condition
-- has is a boolean that indicates whether you want the target to have the condition
-- true - the target will have the condition
-- false - the target will not have the condition

function targetHasCondition(has, condition)
	intruders = nearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		if has and getConditions(intruder) [condition] then
			return intruder
		end
		if not has and not getConditions(intruder) [condition] then
			return intruder
		end
	end
end


-- BUFFing friends. Find friends who need a condition

function allyHasCondition(has, condition)
	allies = nearestNEntities(50, "denizen")
	for _, ally in pairs (allies) do
		if has and getConditions(ally) [condition] then
			return ally
		end
		if not has and not getConditions(ally) [condition] then
			return ally
		end
	end
end




-- How to Chain Targets
--target = pursue()
--if not target then
--	target = retaliate()
--end
--if not target then 
--	target = whatevs()
--end


--Determine extra distance

--Attackstat - current AP


function ApNeeded(attack)
	apCur = getEntityStats(me()) ["apCur"]
	apCost = getBasicAttackStats(me(), attack) ["ap"]
	extra_dist = apCur - apCost
	return extra_dist
end


-- --target and execute AOE attack



-- function AOEtargetAndAttack(attack, extra_dist, spec)
-- 	target = bestAoeAttackPos(attack, extra_dist, spec)
-- 		-- pos = ???
-- 		return target
-- 	-- 	end
-- 	-- if exists(target) then
-- 	-- 	res = doAoeAttack(target, pos)
-- 	-- end
-- end
