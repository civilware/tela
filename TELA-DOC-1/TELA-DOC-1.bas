//  Copyright 2024. Civilware. All rights reserved.
//  TELA Decentralized Web Document (TELA-DOC-1)

Function InitializePrivate() Uint64
10 IF init() == 0 THEN GOTO 30
20 RETURN 1
30 STORE("nameHdr", "<nameHdr>")
31 STORE("descrHdr", "<descrHdr>")
32 STORE("iconURLHdr", "<iconURLHdr>")
33 STORE("dURL", "<dURL>")
34 STORE("docType", "<language>")
35 STORE("subDir", "")
36 STORE("fileCheckC", "<c>")
37 STORE("fileCheckS", "<s>")
100 RETURN 0
End Function

Function init() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("owner", address())
50 STORE("docVersion", "1.0.0") // DOC SC version
60 STORE("hash", HEX(TXID()))
70 STORE("likes", 0)
80 STORE("dislikes", 0)
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

/*
docType code goes in this comment section
*/