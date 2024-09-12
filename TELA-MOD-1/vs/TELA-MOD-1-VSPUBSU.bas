//  Copyright 2024. Civilware. All rights reserved.
//  TELA Decentralized Web Standard Module (TELA-MOD-1)

Function SetVar(k String, v String) Uint64
10 IF LOAD("owner") == "anon" THEN GOTO 20
11 DIM addr as String
12 LET addr = address()
13 IF LOAD("owner") == addr THEN GOTO 30
14 IF addr == "anon" THEN GOTO 20
15 IF STRLEN(k) > 256 || STRLEN(v) > 256 THEN GOTO 20
16 IF EXISTS("var_"+addr+"_"+k) == 0 THEN GOTO 50
20 RETURN 1
30 STORE("var_"+k, v)
40 RETURN 0 
50 STORE("var_"+addr+"_"+k, v)
90 RETURN 0
End Function

Function DeleteVar(k String) Uint64
10 IF EXISTS("var_"+k) == 0 THEN GOTO 20
11 IF LOAD("owner") == "anon" THEN GOTO 20
12 IF LOAD("owner") == address() THEN GOTO 30
20 RETURN 1
30 DELETE("var_"+k)
40 RETURN 0
End Function