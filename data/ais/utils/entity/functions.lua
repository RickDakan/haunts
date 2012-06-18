
-- move to within range and attack target

function moveAndAttack (attack, target)
	ps = allPathablePoints (pos(me()), pos(target), 1, getBasicAttackStats(me(), attack).range)
	res = doMove (ps, 1000)
	if exists(target) then
		doBasicAttack(attack, target)
	end
end
	
--keep distance

function keepDistanceAndAttack (distance, attack, target)
	ps = allPathablePoints (pos(me()), pos(target), distance, getBasicAttackStats(me(), attack).range)
	doMove (ps, 1000)
	if exists(target) then
		doBasickAttack(attack, target)
	end
end





--TARGETING FUNCTIONS
---so, would this function, with the print and return and stuff give me something to play off of

--nearest Target

function adjacent()
	intruders = nearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		dist = rangedDistBetweenEntities(me(), intruder, 1, 10)
		if dist == 1 then
			target = intruder
		end
	end
	return target
end


function pursue()
	print(target)
	intruders = nearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		target = entityInfo(me()).lastEntityIAttacked
	end
	if res == nil then
		return nil
	end
end

function nearest()
	intruders = nearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		target = intruder
	end
	return target
end


function retaliate()
	intruders = nearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		target = entityInfo(me()).lastEntityThatAttackedMe
	end
end

		
--target enemy that attacked nearest ally

function targetAllyAttacker()
	allies = nearestNEntities (50, "denizen")
	for _, ally in pairs (allies) do
		target = entityInfo(ally).lastEntityThatAttackedMe
	end	
	return target
end

--target enemy your allies are already attacking
function targetAllyTarget()
	allies = nearestNEntities (50, "denizen")
	for _, ally in pairs (allies) do
		target = entityInfo(ally).lastEntityThatIAttacked
	end	
	return target
end
	

-- target lowest stat
-- stat is looking for corpus, ego, hpCur, hpMax, apCur, apMax
function targetLowestStat(stat)
	intruders = nearestNEntities (10, "intruder")
	min = 1000
	for _, intruder in pairs (intruders) do
		if GetEntityStats(intruder) [stat] < min then
			min = GetEntityStats(intruder) [stat]
			target = intruder
		end
	end
	return target, min
end

--- Target Highest Stat
-- HOW?

function targetHighestStat(stat)
	intruders = nearestNEntities (10, "intruder")
	min = 1
	for _, intruder in pairs (intruders) do
		if GetEntityStats)intruder) [stat] > min then
			min = GetEntityStats(intruder) [stat]
			target = intruder
		end
	end
	return target, min
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
	for =, ally in pairst (allies) do
		if has and getConditions(ally) [condition] then
			return intruder
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
	apCur = getEntityStats(me()) [apCur]
	apCost = getBasicAttackStats(me()), attack) [ap]
	extra-dist = apCur - apCost
	return extra_dist
end


--target and execute AOE attack



function AOEtargetAndAttack(attack, extra_dist, spec)
	target = bestAoeAttackPos(attack, extra_dist, spec)
		-- pos = ???
		return target
		end
	if exists(target) then
		res = doAoeAttack(target, pos)
	end
end
