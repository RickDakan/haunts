There are four different kinds of ais, each is used for a different purpose and may or may not actually control an individual entity.
- Denizens: This Ai is used as a replacement for a human playing the denizens.
- Intruders: This Ai is used as a replacement for a human playing the intruders.
- Minions: This Ai is used to choose what order to activate all minion-level denizens, regardless of whether the denizens is being controlled by a human or an ai.
- Entity: This Ai is used to control a specific entity.  This is the only ai that can execute actions.

Since the Entity Ai is the only kind that actually executes actions, the rest of the Ais are simply for choosing the order in which to allow entities to execute.  Denizen, Intruder, and Minion Ais are all quite similar, they have access to different functions to prevent them from inadvertently accessing an entity they shouldn't be allowed to access.  For the purposes of these Ais an 'Active' Ai is one which has not completed its actions for the turn.  All Ais are marked as active at the beginning of the turn and are responsible for marking themselves as inactive by way of the 'done' function.  Once an entity is marked as inactive it cannot be controlled again for the rest of the turn.  Being inactive and having 0 Ap are not the same, although it does make sense to call 'done' if you have 0 Ap.

###Denizens Ai
The Denizens Ai must choose the order in which all master and servitor denizen entities execute.  It executes after all minion entities have executed and cannot control them.

**numActiveServitors () -> (float)**  
Returns the number of servitor level denizens that are still active.

**randomActiveServitor () -> (Entity)**  
Returns a random active servitor level denizen.

**exec (Entity) -> ()**  
Calls the Ai for the specified Entity and allows it to execute a single action.

**done () -> ()**  
Indicates that this Ai has completed everything it plans to do this turn.


###Minion Ai
The Minion Ai must choose the order in which all minion level denizens execute.  It executes at the beginning of the turn, before any servitors or masters.

**numActiveMinions () -> (number)**  
Returns the number of minion level denizens that are still active.

**randomActiveMinions () -> (Entity)**  
Returns a random active minion level denizen.

**exec (Entity) -> ()**  
Calls the Ai for the specified Entity and allows it to execute a single action.

**done () -> ()**  
Indicates that this Ai has completed everything it plans to do this turn.


###Entity Ai
An Entity Ai specifies how a single entity functions.  When an Entity Ai is executing it will pause after it takes an action, at which point it may continue executing, or other entities will execute.  When it resumes it should not be assumed that the state of the game is the same as it was when it paused.  For example if an entity commits to move next to an enemy unit, another entity may go next and kill that unit it just stood next to, when it resumes that unit will no longer be there.  If an Ai tries to do something illegal, such as any operation on an entity that does not exist, the Ai will terminate for that turn and an error will be printed to the log.

**me () -> (Entity)**  
Returns the entity of this Ai.

**roll (N,K) -> (number)**  
Does the roll NdK and returns the result.

**numVisibleEnemies () -> (number)**  
Returns the number of enemies that this entity can see.

**nearestEnemy () -> (Entity)**  
Returns the nearest enemy entity to this entity.  Such an entity may not exist, before using this you should check that numVisibleEnemies > 0, or you should check that this entity exists with stillExists.

**distBetween (Entity, Entity) -> (number)**  
Returns the straight-line distance between two entities.

**stillExists (Entity) -> (bool)**  
Returns true iff the entity exists.

**lastOffensiveTarget () -> (Entity)**  
Returns the last entity that this entity attacked.  Note: This entity may not be in LoS and it may not be possible to path to it.

**advanceAllTheWay (Entity) -> ()**  
Advances this entity as far as possible towards the specified entity.

**advance (target Entity, dist number, max number) -> ()**  
Advances this entity towards target until it is within walking-distance dist of it.  While doing so it will not spend more than max Ap.

**moveInRange (Entities, min_dist number, max_dist number, max_ap number) -> ()**  
Advances this entity as short a distance as possible such that it is within min_dist and max_dist walking distance of the specified entities.  While doing so it will not spend more than max Ap.  Note that this function takes an array of entities, so you cannot just pass it a single entity (I need to do something about this).

**costToMoveInRange (Entities, min_dist number, max_dist number) -> (number)**  
Returns the Ap required for this entity to execute moveInRange with these same parameters.

**allIntruders () -> (Entities)**  
Returns an array of all intruders.

**getBasicAttack () -> (string)**  
Returns the name of a random basic attack belonging to this entity.  If this entity does not have a basic attack it will return an empty string.

**doBasicAttack (target Entity, attack string) -> ()**  
Performs the specified attack against the specified target.

**basicAttackStat (attack string, stat string) -> (number)**  
Queries the specified basic attack for one of its stats.  Acceptable values of stat are:
* ap
* damage
* strength
* range
* ammo

**aoeAttackStat (attack string, stat string) -> (number)**  
Queries the specified aoe attack for one of its stats.  Acceptable values of stat are:
* ap
* damage
* strength
* range
* ammo
* diameter

**corpus (Entity) -> (number)**  
Returns the corpus of the specified entity.

**ego (Entity) -> (number)**  
Returns the ego of the specified entity.

**hpMax (Entity) -> (number)**  
Returns the max hp of the specified entity.

**apMax (Entity) -> (number)**  
Returns the max ap of the specified entity.

**hpCur (Entity) -> (number)**  
Returns the current hp of the specified entity.

**apCur (Entity) -> (number)**  
Returns the current ap of the specified entity.

**hasCondition (Entity, condition string) -> (bool)**  
Returns true iff the specified entity currently as the specified condition
