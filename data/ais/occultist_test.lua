-- Check for Ammo on First Aid

function Think()
	intruders = NearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		if intruder.HpCur <6 and rangedDistBetweenEntites (Me(), intruder) <5 then
			if Me().Actions["aid"].Ammo > 0 then
				moveWithinRangeAndAttack (1, "Aid", intruder)
			end
		end
	end
	target = pursue()
	if target == nil then
		target = targetAllyTarget()
	end
	if target == nil then
		target = targetHighestStat("HpCur")
	end
	if getEntityStat(target).Corpus > 9 then
		MoveWithinRangeAndAttack(2, "Exorcise", target)
	else
		MoveWithinRangeAndAttack (2, "Dire Curse", target)
	end
end
	
--Targetting AOE Abjuration here, looking for at least two dudes together, unless Master present
























