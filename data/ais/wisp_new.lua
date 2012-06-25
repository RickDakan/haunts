-- wisp new
-- look at moving away from everyone


function think()
	intruder = nearestNEntities (1, "intruder")[1]
	if rangedDistBetweenEntities (me(), intruder) <3 then
		moveWithinRangeAndAttack (4, "Ectoplasmic Discharge", intruder)
	else 
		moveWithinRangeAndAttack (4, "Ectoplasmic Discharge", intruder)
	end
end
think()