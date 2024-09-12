//  Copyright 2024. Civilware. All rights reserved.
//  TELA Decentralized Web Standard Module (TELA-MOD-1)

Function DepositAsset(scid String) Uint64
10 IF LOAD("owner") == "anon" THEN GOTO 20
11 IF ASSETVALUE(HEXDECODE(scid)) > 0 THEN GOTO 30
20 RETURN 1
30 IF EXISTS("balance_"+scid) THEN GOTO 60
40 STORE("balance_"+scid, ASSETVALUE(HEXDECODE(scid)))
50 GOTO 90
60 STORE("balance_"+scid, LOAD("balance_"+scid)+ASSETVALUE(HEXDECODE(scid)))
90 RETURN 0
End Function

Function WithdrawAsset(scid String, amt Uint64) Uint64
10 IF LOAD("owner") == "anon" THEN GOTO 20
12 IF EXISTS("balance_"+scid) == 0 THEN GOTO 20
13 IF LOAD("balance_"+scid) < amt THEN GOTO 20
14 IF LOAD("owner") == address() THEN GOTO 30
20 RETURN 1
30 SEND_ASSET_TO_ADDRESS(SIGNER(), amt, HEXDECODE(scid))
40 STORE("balance_"+scid, LOAD("balance_"+scid)-amt)
90 RETURN 0
End Function