//  Copyright 2024. Civilware. All rights reserved.
//  TELA Decentralized Web Standard (TELA-INDEX-1)

Function InitializePrivate() Uint64
10 IF init() == 0 THEN GOTO 30
20 RETURN 1
30 STORE("nameHdr", "<nameHdr>")
31 STORE("descrHdr", "<descrHdr>")
32 STORE("iconURLHdr", "<iconURLHdr>")
33 STORE("dURL", "<dURL>")
34 STORE("mods", "<modTags>")
40 STORE("DOC1", "<scid>") 
// 41 STORE("DOC2", "<scid>")
// 42 STORE("DOC3", "<scid>")
1000 RETURN 0
End Function

Function init() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("owner", address())
50 STORE("telaVersion", "1.1.0") // TELA SC version
60 STORE("commit", 0) // The initial commit
70 STORE(0, HEX(TXID())) // SCID commit hash
80 STORE("hash", HEX(TXID()))
85 STORE("likes", 0)
90 STORE("dislikes", 0)
100 RETURN 0
End Function

Function address() String
10 DIM s as String
20 LET s = SIGNER()
30 IF IS_ADDRESS_VALID(s) THEN GOTO 50
40 RETURN "anon"
50 RETURN ADDRESS_STRING(s) 
End Function

Function Rate(r Uint64) Uint64
10 DIM addr as String
15 LET addr = address()
16 IF r < 100 && EXISTS(addr) == 0 && addr != "anon" THEN GOTO 30
20 RETURN 1
30 STORE(addr, ""+r+"_"+BLOCK_HEIGHT())
40 IF r < 50 THEN GOTO 70
50 STORE("likes", LOAD("likes")+1)
60 RETURN 0
70 STORE("dislikes", LOAD("dislikes")+1)
100 RETURN 0
End Function

Function UpdateCode(code String, mods String) Uint64
10 IF LOAD("owner") == "anon" THEN GOTO 20
15 IF code == "" THEN GOTO 20
16 IF LOAD("owner") == address() THEN GOTO 30
20 RETURN 1
30 UPDATE_SC_CODE(code)
40 STORE("commit", LOAD("commit")+1) // New commit
50 STORE(LOAD("commit"), HEX(TXID())) // New hash
60 STORE("hash", HEX(TXID()))
70 STORE("mods", mods)
100 RETURN 0
End Function