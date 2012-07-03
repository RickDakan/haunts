-- wisp new
-- look at moving away from everyone


function Think()
	intruder = nearestNEntities (1, "intruder")[1]
	if rangedDistBetweenEntities (Me, intruder) <3 then
		moveWithinRangeAndAttack (4, "Ectoplasmic Discharge", intruder)
	else 
		moveWithinRangeAndAttack (4, "Ectoplasmic Discharge", intruder)
	end
end
