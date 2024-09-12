//  Copyright 2024. Civilware. All rights reserved.
//  TELA Decentralized Web Standard Module (TELA-MOD-1)

Function TransferOwnership(addr String) Uint64  
10 IF LOAD("owner") == "anon" THEN GOTO 20
11 IF IS_ADDRESS_VALID(ADDRESS_RAW(addr)) == 0 THEN GOTO 20
12 IF LOAD("owner") == address() THEN GOTO 30 // address() comes from TELA base code
20 RETURN 1
30 STORE("tmpowner", addr) // Use the string to match address()
40 RETURN 0
End Function

Function ClaimOwnership() Uint64
10 IF LOAD("owner") == "anon" THEN GOTO 20
11 IF EXISTS("tmpowner") == 0 THEN GOTO 20
12 IF LOAD("tmpowner") == address() THEN GOTO 30
20 RETURN 1
30 STORE("owner", address())
40 DELETE("tmpowner")
50 RETURN 0
End Function