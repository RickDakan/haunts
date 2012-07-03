--Teen
-- add in use Telepathic Shroud when near Goal
--UTIL Entity Folders - to alls
-- also WILL have Minions, Denizens, Intruders added to 

function Think()
	if Me.HpCur < 5 and not Me.Conditions["Psychic Shroud"] then
		doBasicAttack ("Psychic Shroud", Me)
	end
	denizens = nearestNEntities (50, "denizen")
	for _, denizen in pairs (denizens) do
		if denizen.Corpus >10 and not denizen.Conditions["Telepathic Target"] then
			MoveWithinRangeAndAttack (1, "Telepathic Coordination", denizen)
		end
	end
	target = Pursue()
	if target == nil then
		target = retaliate()
	end
	if target == nil then
		target = targetAllyTarget()
	end
	if target == nil then
		target = targetLowestStat("HpCur")
	end
	if target == nil then
		return
	end
	if rangedDistBetweenEntities (Me, target) <2 then
		MoveWithinRangeAndAttack(1, "Kick", target)
	else
		MoveWithinRangeAndAttack (2, "Pistol", target)
	end
end

			
		