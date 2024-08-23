// Global websocket
let socket;

// XSWD application data
const applicationData = {
    "id": "71605a32e3b0c44298fc1c549afbf4c8496fb92427ae41e4649b934ca495991b",
    "name": "TELA Demo Application",
    "description": "Basic WS connection parts for TELA application",
    "url": "http://localhost:" + location.port, // Get the current port being used by server to set in our XSWD application data, must match origin URL
};

let typed = 0;
let typeText = "";
const typeSpeed = 50;

const jsonBody = document.getElementById("jsonDisplayBody");
const jsonResult = document.getElementById("jsonDisplayResult");

// Typing text effect
function typeWriter(text) {
    const html = document.getElementById("typingLabel");
    if (typed === 0) {
        html.innerHTML = "";
        typeText = text;
    }

    if (typed < typeText.length) {
        html.innerHTML += typeText.charAt(typed);
        typed++;
        setTimeout(typeWriter, typeSpeed);
    }

    if (typed === typeText.length) {
        setTimeout(() => {
            typed = 0;
        }, typeSpeed);
    }
}

// Function to send data
function sendData(d) {
    if (socket && socket.readyState === WebSocket.OPEN) {
        try {
            socket.send(JSON.stringify(d));
            if (d.method) {
                console.log(d.method, "request sent to the server");
                if (jsonBody) jsonBody.innerHTML = JSON.stringify(d, null, 2);
                if (jsonResult) jsonResult.innerHTML = "";
            } else {
                console.log("Connection request sent to the server");
            }
        } catch (error) {
            console.error("Failed to send data:", error);
            toggleIndicators("red");
        }
    } else {
        console.log("Web socket is not open. State:", socket ? socket.readyState : "N/A");
        toggleIndicators("red");
    }
}

// Handle web socket connection and listeners
function connectWebSocket() {
    // If we are already connected, disconnect
    if (document.getElementById("connectButton").textContent === "Disconnect") {
        toggleIndicators("red");
        return;
    }

    // Connect to the web socket
    socket = new WebSocket("ws://localhost:44326/xswd");

    // Listen for open
    socket.addEventListener("open", function (event) {
        console.log("Web socket connection established:", event);
        toggleIndicators("yellow");
        typed = 0;
        typeWriter("Waiting for wallet reply");
        sendData(applicationData); // Send ApplicationData after connection is established
    });

    let address = "";
    let connecting = true;

    // Listen for the messages
    socket.addEventListener("message", function (event) {
        const response = JSON.parse(event.data);
        console.log("Response received:", response);
        if (response.accepted) { // If connection is accepted, we will request to get address from wallet
            console.log("Connected message received:", response.message);
            sendData({
                jsonrpc: "2.0",
                id: "1",
                method: "GetAddress"
            });
        } else if (response.result) {
            const res = response.result;
            if (jsonResult) jsonResult.innerHTML = JSON.stringify(res, null, 2);
            typed = 0;
            if (res.address) { // If GetAddress is allowed by user
                address = res.address;
                console.log("Connected address:", address);
                toggleIndicators("green");
                connecting = false;
                typeWriter(address);
            } else if (res.unlocked_balance) { // Balance response
                const bal = "Balance: " + (res.unlocked_balance / 100000).toFixed(5) + " DERO";
                console.log(bal);
                typeWriter(bal);
            } else if (res.height) { // Height response
                const height = "Height: " + res.height;
                console.log(height);
                typeWriter(height);
            } else if (res.telaLinkResult) {  // TELA link response
                const link = "Opened TELA link: " + res.telaLinkResult;
                console.log(link);
                typeWriter(link);
            } else if (res.lastIndexHeight) { // Gnomon responses
                const gnomon = "Gnomon indexed height: " + res.lastIndexHeight;
                console.log(gnomon);
                typeWriter(gnomon);
            } else if (res.epochHashes) {  // EPOCH responses
                const epoch = "Hashes: " + res.epochHashes + " in " + res.epochDuration + "ms and submitted " + res.epochSubmitted + " as miniblocks";
                console.log(epoch);
                typeWriter(epoch);
            } else if (res.epochAddress) {
                const epoch = "EPOCH address: " + res.epochAddress;
                console.log(epoch);
                typeWriter(epoch);
            } else if (res.maxHashes) {
                const epoch = "EPOCH max hashes: " + res.maxHashes;
                console.log(epoch);
                typeWriter(epoch);
            } else if (res.sessionMinis) {
                const epoch = "EPOCH session hashes: " + res.sessionHashes + "  miniblocks: " + res.sessionMinis;
                console.log(epoch);
                typeWriter(epoch);
            }
        } else if (response.error) { // Display error message
            console.error("Error:", response.error.message);
            typed = 0;
            typeWriter(" " + response.error.message);
            document.getElementById("typingLabel").innerHTML = "";
            if (connecting) toggleIndicators("red");
        }
    });

    // Listen for errors
    socket.addEventListener("error", function (event) {
        console.error("Web socket error:", event);
        typed = 0;
        typeWriter(" Web socket error: " + event.target.url.toString());
    });

    // Listen for close
    socket.addEventListener("close", function (event) {
        console.log("Web socket connection closed:", event.code, event.reason);
        toggleIndicators("red");
    });
}

// Change indictor color based on connection status
function toggleIndicators(color) {
    if (color === "green") {
        document.getElementById("connectButton").textContent = "Disconnect";
        document.getElementById("greenIndicator").style.display = "block";
        document.getElementById("yellowIndicator").style.display = "none";
        document.getElementById("redIndicator").style.display = "none";
    } else if (color === "yellow") {
        document.getElementById("connectButton").textContent = "Disconnect";
        document.getElementById("greenIndicator").style.display = "none";
        document.getElementById("yellowIndicator").style.display = "block";
        document.getElementById("redIndicator").style.display = "none";
    } else {
        document.getElementById("connectButton").textContent = "Connect";
        document.getElementById("greenIndicator").style.display = "none";
        document.getElementById("yellowIndicator").style.display = "none";
        document.getElementById("redIndicator").style.display = "block";
        document.getElementById("typingLabel").textContent = "";
        if (socket) socket.close(), socket = null;
    }
}

// Create new text input
function newInput(p, n) {
    const input = document.createElement("input");
    if (n) {
        input.type = "number";
        input.step = 1;
        input.min = 0;
    } else {
        input.type = "text";
    }

    input.id = p;
    input.placeholder = p + ":";

    return input;
}

// Create needed params for request
function requestParams() {
    const container = document.getElementById("paramsContainer");
    const select = document.getElementById("selectCall");
    container.innerHTML = "";
    switch (select.value) {
        case "HandleTELALinks":
            container.appendChild(newInput("telaLink")); break;
        case "GetTxCount":
            container.appendChild(newInput("txType")); break;
        case "GetOwner":
        case "GetAllNormalTxWithSCIDBySCID":
        case "GetAllSCIDInvokeDetails":
        case "GetAllSCIDVariableDetails":
        case "GetSCIDInteractionHeight":
            container.appendChild(newInput("scid")); break;
        case "GetAllNormalTxWithSCIDByAddr":
        case "GetMiniblockCountByAddress":
        case "GetSCIDInteractionByAddr":
            container.appendChild(newInput("address")); break;
        case "GetAllSCIDInvokeDetailsByEntrypoint":
            container.appendChild(newInput("scid"));
            container.appendChild(newInput("entrypoint")); break;
        case "GetAllSCIDInvokeDetailsBySigner":
            container.appendChild(newInput("scid"));
            container.appendChild(newInput("signer")); break;
        case "GetSCIDVariableDetailsAtTopoheight":
            container.appendChild(newInput("scid"));
            container.appendChild(newInput("height", true)); break;
        case "GetSCIDKeysByValue":
        case "GetSCIDValuesByKey":
            container.appendChild(newInput("scid"));
            container.appendChild(newInput("height", true));
            if (select.value === "GetSCIDKeysByValue") {
                container.appendChild(newInput("value"));
            } else {
                container.appendChild(newInput("key"));
            }
            break;
        case "GetLiveSCIDValuesByKey":
        case "GetLiveSCIDKeysByValue":
            container.appendChild(newInput("scid"));
            if (select.value === "GetLiveSCIDKeysByValue") {
                container.appendChild(newInput("value"));
            } else {
                container.appendChild(newInput("key"));
            }
            break;
        case "GetInteractionIndex":
            container.appendChild(newInput("topoheight", true));
            container.appendChild(newInput("height", true)); break;
        case "GetMiniblockDetailsByHash":
            container.appendChild(newInput("blid")); break;
        case "AttemptEPOCH":
            container.appendChild(newInput("hashes", true)); break;
    }
}

// Send call with params
function callFor() {
    if (!socket) {
        typed = 0;
        typeWriter("Wallet is not connected");
        return;
    }

    typed = 0;
    document.getElementById("typingLabel").innerHTML = "";

    const select = document.getElementById("selectCall");
    let m = "";
    if (select.selectedIndex > 3 && !select.value.endsWith("EPOCH")) {
        m = "Gnomon.";
    } else {
        typeWriter("Waiting for wallet reply");
    }

    const call = {
        jsonrpc: "2.0",
        id: "1",
        method: m + select.value,
        params: {},
    };

    switch (select.value) {
        case "HandleTELALinks":
            call.params.telaLink = document.getElementById("telaLink").value; break;
        case "GetTxCount":
            call.params.txType = document.getElementById("txType").value; break;
        case "GetOwner":
        case "GetAllNormalTxWithSCIDBySCID":
        case "GetAllSCIDInvokeDetails":
        case "GetAllSCIDVariableDetails":
        case "GetSCIDInteractionHeight":
            call.params.scid = document.getElementById("scid").value; break;
        case "GetAllNormalTxWithSCIDByAddr":
        case "GetMiniblockCountByAddress":
        case "GetSCIDInteractionByAddr":
            call.params.address = document.getElementById("address").value; break;
        case "GetAllSCIDInvokeDetailsByEntrypoint":
            call.params.scid = document.getElementById("scid").value;
            call.params.entrypoint = document.getElementById("entrypoint").value; break;
        case "GetAllSCIDInvokeDetailsBySigner":
            call.params.scid = document.getElementById("scid").value;
            call.params.signer = document.getElementById("signer").value; break;
        case "GetSCIDVariableDetailsAtTopoheight":
            call.params.scid = document.getElementById("scid").value;
            call.params.height = parseInt(document.getElementById("height").value); break;
        case "GetSCIDKeysByValue":
        case "GetSCIDValuesByKey":
            call.params.scid = document.getElementById("scid").value;
            call.params.height = parseInt(document.getElementById("height").value);
            if (select.value === "GetSCIDKeysByValue") {
                call.params.value = document.getElementById("value").value;
            } else {
                call.params.value = document.getElementById("key").value;
            }
            break;
        case "GetLiveSCIDValuesByKey":
        case "GetLiveSCIDKeysByValue":
            call.params.scid = document.getElementById("scid").value;
            if (select.value === "GetLiveSCIDKeysByValue") {
                call.params.value = document.getElementById("value").value;
            } else {
                call.params.value = document.getElementById("key").value;
            }
            break;
        case "GetInteractionIndex":
            call.params.topoheight = parseInt(document.getElementById("topoheight").value);
            call.params.height = parseInt(document.getElementById("height").value); break;
        case "GetMiniblockDetailsByHash":
            call.params.blid = document.getElementById("blid").value; break;
        case "AttemptEPOCH":
            call.params.hashes = parseInt(document.getElementById("hashes").value); break;
        default:
            call.params = null; break;
    }

    sendData(call);
}