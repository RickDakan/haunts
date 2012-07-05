

-- if there's someone next to him without Agony, inject him
-- move away from someone
-- shoot them!


--check to see if adjacent people have Agony - if they do, he wants to move awa

function Think()
	intruders = Utils.NearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		if Utils.RangedDistBetweenEntities (Me, intruder) <2 then
			if intruder.Conditions ["Agony"] then
				moveWithinRangeAndAttack(3, "Envenomed Dart", intruder)
			else
				moveWithinRangeAndAttack (1, "Inject", intruder)
			end
		end
		target = pursue()
		if target == nil then
			target = targetAllyTarget()
		end
		if target == nil then
			target = targetAllyAttacker()
		end
		if target == nil then
			target = retaliate()
		end
		if target == nil then
			target = targetLowestStat("Ego")
		end
		if target == nil then
			target = nearest()
		end	
		if target == nil then
			return
		end
	--	if target.Conditions["Agony"] then
		moveWithinRangeAndAttack (3, "Envenomed Dart", target)
	--	else
	--		moveWithinRangeAndAttack (1, "Inject", target)
	end
end
