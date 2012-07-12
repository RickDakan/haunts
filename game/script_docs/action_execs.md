Action Exec Objects
-------------------

A script's OnAction() function is called after every action.

###__OnAction__(_intruders_, _round_, _exec_)
_intruders_: True iff it is the intruders' turn.  
_round_: What round it is.  
_exec_: The Action Exec object describing the action that just took place.  

------

All exec objects have the following two fields:  

_Action_: The Action object describing the action that just happened.  
_Ent_: The entity that performed the action.  

Each kind of action has fields specific to it:  

####Basic Attacks
_Target_: The entity that was targeted by the action.  

------

####Aoe Attacks
_Pos_: The center of the aoe.  

------

####Move Actions
_Path_: An array of points indicating what positions the entity moved through.  

------

####Interact Actions
_Toggle Door_: A boolean indicating if this action was to open/close a door.  
If _Toggle Door_ is true then the exec object will also have the following fields:  
  _Door_: The door that was open/closed.  
else  
  _Target_: The entity that was targeted by the action.  

------

####Summon Actions
_Pos_: The position the entity was summoned to.  

