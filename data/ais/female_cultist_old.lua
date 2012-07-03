-- ALL THESE FUNCTIONS WOULD BE UTILITIES
-- 	But I put them here for reference right now 

function moveAndAttack (attack, target)
	ps = allPathablePoints (Me.Pos, pos(target), 1, getBasicAttackStats(Me, attack).range)
	res = doMove (ps, 1000)
	if exists(target) then
		doBasicAttack(attack, target)
	end
end
	
function pursue()
	print(target)
	intruders = nearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		target = entityInfo(Me).lastEntityIAttacked
	end
	if res == nil then
		return nil
	end
	return target
end

function retaliate()
	intruders = nearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		target = entityInfo(Me).lastEntityThatAttackedMe
	end
	return target
end


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
			min = GetEntityStatsFunc(intruder) [stat]
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
  



--The Female Cultist AI would begin here, assuming all the above is in the Utility folder

function Think()
	target = pursue()
	if not target then
		target = retaliate()
	end
	if not target then
		target = targetAllyAttacker()
	end
	if not target then
		target = targetAllyTarget()
	end
	if not target then
		target = targetLowestState(hpCur)
	end

	if targetHasCondition(true, "Blindness") then
		moveAndAttack ("Envenomed Blade", target)
	end
	if targetHasCondition(false, "Blindness") then
		moveAndAttack ("Parasitic Gift", target)
	end
end
