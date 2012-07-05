-- wisp new
-- look at moving away from everyone


function Think()
	intruder = Utils.NearestNEntities (1, "intruder")[1]
	if Utils.RangedDistBetweenEntities (Me, intruder) <3 then
		moveWithinRangeAndAttack (4, "Ectoplasmic Discharge", intruder)
	else 
		moveWithinRangeAndAttack (4, "Ectoplasmic Discharge", intruder)
	end
end
