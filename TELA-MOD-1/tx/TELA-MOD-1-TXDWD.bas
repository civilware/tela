//  Copyright 2024. Civilware. All rights reserved.
//  TELA Decentralized Web Standard Module (TELA-MOD-1)

Function DepositDero() Uint64
10 IF LOAD("owner") == "anon" THEN GOTO 20
11 IF DEROVALUE() > 0 THEN GOTO 30
20 RETURN 1
30 IF EXISTS("balance_dero") THEN GOTO 60
40 STORE("balance_dero", DEROVALUE())
50 GOTO 90
60 STORE("balance_dero", LOAD("balance_dero")+DEROVALUE())
90 RETURN 0
End Function

Function WithdrawDero(amt Uint64) Uint64
10 IF LOAD("owner") == "anon" THEN GOTO 20
12 IF EXISTS("balance_dero") == 0 THEN GOTO 20
13 IF LOAD("balance_dero") < amt THEN GOTO 20
14 IF LOAD("owner") == address() THEN GOTO 30
20 RETURN 1
30 SEND_DERO_TO_ADDRESS(SIGNER(), amt)
40 STORE("balance_dero", LOAD("balance_dero")-amt)
90 RETURN 0
End Function