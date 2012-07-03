--master AI
--spec
--"allies ok", "minions ok", "enemies only"

--no enemies in sight?
	-- look at shades and wisps,
	-- whichever I see fewer of, summon one.

--needs a target?
function aoePlaceAndAttack(attack, spec)
	gz = bestAoeAttackPos (attack, Me.ApCur - Me.Actions[attack].Ap, spec)
	dsts = AllPathablePoints(Me.Pos, gz, 1, Me.Actions[attack].Range)
	doMove(dsts, 1000)
	if RangedDistBetweenPositions (Me.Pos, gz) > Me.Actions[attack].Range then
		return
	else
		doAoeAttack(attack, gz)
	end
end

function Think()
	aoePlaceAndAttack("Poltergeist Blast", "enemies only")
	Think()
end

