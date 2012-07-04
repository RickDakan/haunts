-- cult leader

-- buffs with great power
-- strongest ally
-- not if enemies within 5 of him.
-- alternate buff and debuff or buff first
-- buff, then debuff, then retreat our of range and attack

--


function Think()
	intruders = NearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		if RangedDistBetweenEntities (Me(), intruder) <3 then
			moveWithinRangeAndAttack(3, "Crozier", intruder)
		else
		target = allyHasCondition(false, "Inspired")
			if target ~= nil then
				moveWithinRangeAndAttack (1, "Voice of the Beyond", target)
			else
				target = targetHasCondition(false, "Panic")
				if target ~= nil then
					moveWithinRangeAndAttack (5, "Revelations of Despair", target)
				else
					target = targetLowestStat("Ego")
						moveWithinRangeAndAttack (5, "Crozier", target)
				end
			end
		end
		Think()
	end
end

