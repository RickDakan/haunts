--wraith

function think()
	target = targetHasCondition(false, "Horrified")
	if target ~= nil then	
		moveWithinRangeAndAttack (3, "Grave Grasp", target)
	else 
		target = targetHasCondition(false, "Dread")
		if target ~= nil then
			moveWithinRangeAndAttack (3, "Vengeful Curse", target)
		else
			target = allyHasCondition(false, "Focused")
			if target ~= nil then
				moveWithinRangeAndAttack (1, "Ghastly Howl", target)
			end
		end
	end
	think()
end
think()

		