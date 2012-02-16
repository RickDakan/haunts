package game

import (
  "math/rand"
  "github.com/runningwild/haunts/game/status"
)

func DoAttack(attacker,defender *Entity, strength int, kind status.Kind) bool {
  // get attacker's bonus for using the specified kind of attack
  // get defender's bonus for defending against the specified kind of attack
  // get the defender's current ego/corpus
  // successful attack = strength + attack bonus + 1d10 >= defense bonus + ego/corpus
  attack := attacker.Stats.AttackBonusWith(kind)
  defense := defender.Stats.DefenseVs(kind)
  roll := rand.Intn(10) + 1
  return strength + attack + roll >= defense
}
