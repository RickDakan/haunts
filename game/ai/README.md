There are four different kinds of ais, each is used for a different purpose and may or may not actually control an individual entity.
- Denizens: This Ai is used as a replacement for a human playing the denizens.
- Intruders: This Ai is used as a replacement for a human playing the intruders.
- Minions: This Ai is used to choose what order to activate all minion-level denizens, regardless of whether the denizens is being controlled by a human or an ai.
- Entity: This Ai is used to control a specific entity.  This is the only ai that can execute actions.

Since the Entity Ai is the only kind that actually executes actions, the rest of the Ais are simply for choosing the order in which to allow entities to execute.  Denizen, Intruder, and Minion Ais are all quite similar, they have access to different functions to prevent them from inadvertently accessing an entity they shouldn't be allowed to access.  For the purposes of these Ais an 'Active' Ai is one which has not completed its actions for the turn.  All Ais are marked as active at the beginning of the turn and are responsible for marking themselves as inactive by way of the 'done' function.  Once an entity is marked as inactive it cannot be controlled again for the rest of the turn.  Being inactive and having 0 Ap are not the same, although it does make sense to call 'done' if you have 0 Ap.

All Ais have access to the following mathematical operators and functions:
    +, -, *, /, ^, ln, log2, log10, abs, <, <=, >, >=, ==, pi, e

As well as the following boolean operators:
    && - logical and
    || - logical or
    ^^ - logical xor
    !  - logical not


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

**walkingDistBetween (Entity, Entity) -> (number)**  
Returns the distance one entity would need to walk to be standing in the other's position.  Includes navigating around furniture and other entities.

**rangedDistBetween (Entity, Entity) -> (number)**  
Returns the straight-line distance between two entities, this is the distance used when determining if an entity is in range of an attack.

**stillExists (Entity) -> (bool)**  
Returns true iff the entity exists.

**lastEntAttackedBy (Entity) -> (Entity)**  
Returns the entity that was last attacked by this entity.  If there were multiple (as in the case of an aoe attack) it only returns one of them.  If it returns an entity that entity will be an enemy.

**lastEntThatAttacked (Entity) -> (Entity)**  
Returns the last entity that attacked this one.

**group (Entity) -> ([]Entity)**  
Create an array of entities consiting of only the specified entity.  Useful when using functions that take an array of entities when you only have a single entity.

**advanceInRange ([]Entity, min number, max number, ap number) -> ()**  
Advances the entity along the shortest path that ends at least max distance from an entity in the array and not closer than min distance to any entity in the array.  Regardless of the path or final position, this ent will not spend more than the specified amount of ap in making this move.

**costToMoveInRange ([]Entity, min number, max number) -> (number)**  
Returns the Ap required for this entity to execute advanceInRange with these same parameters.

**advanceAllTheWay ([]Entity) -> ()**  
Advances this entity as far as possible towards one entity in the array.

**moveInRange (Entities, min_dist number, max_dist number, max_ap number) -> ()**  
Advances this entity as short a distance as possible such that it is within min_dist and max_dist walking distance of the specified entities.  While doing so it will not spend more than max Ap.  Note that this function takes an array of entities, so you cannot just pass it a single entity (I need to do something about this).

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

**entityStat (Entity, stat string) -> (number)**  
Queries the specified entity for one of its stats.  Acceptable values of stat are:
* corpus
* ego
* hpMax
* apMax
* hpCur
* apCur

**hasCondition (Entity, condition string) -> (bool)**  
Returns true iff the specified entity currently has the specified condition

**hasAction (Entity, action string) -> (bool)**  
Returns true iff the specified entity currently has the specified action
