function moveWithinRangeAndAttack (min_range, attack, target)
	max_range = Me.Actions[attack].Range
	if min_range > max_range then
		min_range = max_range
	end
	ps = Utils.AllPathablePoints (Me.Pos, target.Pos, min_range, max_range)
	res = Actions.Move (ps, 1000)
	if Utils.Exists(target) then
		Actions.BasicAttack(attack, target)
	end
end

function pursue()
	target = Me.Info.LastEntityThatIAttacked
	if Utils.Exists(target) then
		return target
	end
	return nil
end

function nearest()
	intruders = Utils.NearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		return intruder
	end
	return nil
end


function retaliate()
	target = Me.Info.LastEntityThatAttackedMe
	if Utils.Exists(target) then
		return target
	end
	return nil
end

		
--target enemy that attacked nearest ally

function targetAllyAttacker()
	allies = Utils.NearestNEntities (50, "denizen")
	for _, ally in pairs (allies) do
	  target = ally.Info.LastEntityThatAttackedMe
	  if Utils.Exists(target) then
	  	return target
	  end
	end	
	return nil
end

--target enemy your allies are already attacking
function targetAllyTarget()
	allies = Utils.NearestNEntities (50, "denizen")
	for _, ally in pairs (allies) do
	  target = ally.Info.LastEntityThatIAttacked
	  if Utils.Exists(target) then
	  	return target
	  end
	end	
	return nil
end
	

-- target lowest stat
-- stat is looking for Corpus, Ego, HpCur, HpMax, ApCur, ApMax
function targetLowestStat(stat)
	intruders = Utils.NearestNEntities (10, "intruder")
	target = nil
	min = 10000
	for _, intruder in pairs (intruders) do
		if intruder[stat] < min then
			min = intruder[stat]
			target = intruder
		end
	end
	return target, min
end

--- Target Highest Stat
-- HOW?

function targetHighestStat(stat)
	intruders = Utils.NearestNEntities (10, "intruder")
	max = 0
	for _, intruder in pairs (intruders) do
		if intruder[stat] > max then
			max = intruder[stat]
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
	intruders = Utils.NearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		if has and intruder.Conditions[condition] then
			return intruder
		end
		if not has and not intruder.Conditions[condition] then
			return intruder
		end
	end
end


-- BUFFing friends. Find friends who need a condition

function allyHasCondition(has, condition)
	allies = Utils.NearestNEntities(50, "denizen")
	for _, ally in pairs (allies) do
		if has and ally.Conditions[condition] then
			return ally
		end
		if not has and not ally.Conditions[condition] then
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
	ApCur = Me.ApCur
	ApCost = Me.Actions[attack].Ap
	extra_dist = ApCur - ApCost
	return extra_dist
end


-- --target and execute AOE attack



-- function AOEtargetAndAttack(attack, extra_dist, spec)
-- 	target = Utils.BestAoeAttackPos(attack, extra_dist, spec)
-- 		-- pos = ???
-- 		return target
-- 	-- 	end
-- 	-- if Utils.Exists(target) then
-- 	-- 	res = Actions.AoeAttack(target, pos)
-- 	-- end
-- end
