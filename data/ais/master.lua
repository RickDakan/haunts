--master AI
--spec
--"allies ok", "minions ok", "enemies only"

--no enemies in sight?
	-- look at shades and wisps,
	-- whichever I see fewer of, summon one.

--needs a target?
function aoePlaceAndAttack(attack, spec)
	me_stats = getEntityStats(me())
	attack_stats = getAoeAttackStats(me(), attack)
	gz = bestAoeAttackPos (attack, me_stats.apCur - attack_stats.ap, spec)
	dsts = allPathablePoints(pos(me()), gz, 1, attack_stats.range)
	doMove(dsts, 1000)
	if rangedDistBetweenPositions (pos(me()), gz) > attack_stats.range then
		return
	else
		doAoeAttack(attack, gz)
	end
end

function think()
	aoePlaceAndAttack("Poltergeist Blast", "enemies only")
	think()
end
think()

--	intruders = nearestNEntities (10, "intruder")
	
--	for _, intruder in pairs (intruders) do
--		if rangedDistBetweenEntities (me(), intruder) >1 then
--			aoeTargetAndAttack ("Visions of Despair", "allies ok")
--		else
--			if rangedDistBetweenEntities (me(), intruder) >2 and <6 then
--				aoeTargetAndAttack ("Poltergeist Blast", "minions ok")
--			else
--				aoeTargetAndAttack ("Arc of Decay", "minions ok")
--			end
--		end
--	end
--end
--think()


-- >6 and rangedDistBetweenEntities (me(), intruder) <13 then
		
--if rangedDistBetweenEntities (me(), intruder) >6 and <13 then
--	aoeTargetAndAttack ("Visions of Despair", "allies ok")
--else
--	if rangedDistBetweenEntities (me(), intruder) >2 and <6 then
--		aoeTargetAndAttack ("Poltergeist Blast", "minions ok")
--	else
--		aoeTargetAndAttack ("Arc of Decay", "minions ok")
--	end
--end
	
	