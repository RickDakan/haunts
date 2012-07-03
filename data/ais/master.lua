--master AI
--spec
--"allies ok", "minions ok", "enemies only"

--no enemies in sight?
	-- look at shades and wisps,
	-- whichever I see fewer of, summon one.

--needs a target?
function aoePlaceAndAttack(attack, spec)
	me_stats = getEntityStats(Me)
	attack_stats = getAoeAttackStats(Me, attack)
	gz = bestAoeAttackPos (attack, me_stats.apCur - attack_stats.ap, spec)
	dsts = allPathablePoints(Me.Pos, gz, 1, attack_stats.range)
	doMove(dsts, 1000)
	if rangedDistBetweenPositions (Me.Pos, gz) > attack_stats.range then
		return
	else
		doAoeAttack(attack, gz)
	end
end

function Think()
	aoePlaceAndAttack("Poltergeist Blast", "enemies only")
	Think()
end

