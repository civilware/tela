//  Copyright 2024. Civilware. All rights reserved.
//  TELA Decentralized Web Standard Module (TELA-MOD-1)

Function SetVar(k String, v String) Uint64
10 IF LOAD("owner") == "anon" THEN GOTO 20
11 IF LOAD("owner") == address() THEN GOTO 30
20 RETURN 1
30 STORE("var_"+k, v)
40 RETURN 0 
End Function

Function DeleteVar(k String) Uint64
10 IF EXISTS("var_"+k) == 0 THEN GOTO 20
11 IF LOAD("owner") == "anon" THEN GOTO 20
12 IF LOAD("owner") == address() THEN GOTO 30
20 RETURN 1
30 DELETE("var_"+k)
40 RETURN 0
End Function